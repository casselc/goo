// Isolated startup acceptance: the process cwd contains no lab authoring assets.
const fs = require('node:fs/promises');
const path = require('node:path');
const os = require('node:os');
const crypto = require('node:crypto');
const { spawn } = require('node:child_process');
const base = 'http://127.0.0.1:18088';
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
async function inventory(root, relative = '') {
  const out = [];
  for (const entry of await fs.readdir(path.join(root, relative), { withFileTypes: true })) {
    const name = path.join(relative, entry.name);
    if (entry.isDirectory()) out.push(...await inventory(root, name));
    else out.push({ path: name, sha256: sha(await fs.readFile(path.join(root, name))) });
  }
  return out.sort((a, b) => a.path.localeCompare(b.path));
}
async function startRuntime(binary, artifactDir, options = {}) {
  if (!binary || !path.isAbsolute(binary)) throw Error('GOO_GENERIC_BINARY must name an explicit absolute compiled binary');
  const root = await fs.mkdtemp(path.join(os.tmpdir(), 'goo-generic-191-'));
  await fs.mkdir(artifactDir, { recursive: true });
  await fs.copyFile(path.resolve('server.py'), path.join(root, 'server.py'));
  for (const name of ['ui', 'api']) await fs.cp(path.resolve(name), path.join(root, name), { recursive: true });
  await fs.mkdir(path.join(root, 'captures'));
  if (options.captureStore) await fs.cp(options.captureStore, path.join(root, 'captures'), { recursive: true });
  const files = await inventory(root);
  if (files.some(file => /(^|\/)(grammar\.py|definitions|config|prebuilts|samples)(\/|$)/.test(file.path))) throw Error('Lab assets leaked into isolated runtime');
  const environment = { ...process.env, CLICKHOUSE_URL: options.clickHouseURL || process.env.GOO_TEST_CLICKHOUSE_URL || 'http://127.0.0.1:8123',
    CLICKHOUSE_USER: process.env.GOO_TEST_CLICKHOUSE_USER || 'lab', CLICKHOUSE_PASSWORD: process.env.GOO_TEST_CLICKHOUSE_PASSWORD || 'lab-local',
    GOO_PROFILE: 'generic', GOO_HOST: '127.0.0.1', GOO_PORT: '18088', PACKAGES_URL: 'http://127.0.0.1:18089', PYTHONDONTWRITEBYTECODE: '1' };
  // Explicitly remove optional ambient feature selectors from the acceptance process.
  delete environment.GOO_XTDB_URL;
  delete environment.PYTHONPATH;
  const children = [], commands = [];
  async function launch(command, args, label) {
    const child = spawn(command, args, { cwd: root, env: environment, stdio: ['ignore', 'pipe', 'pipe'] });
    children.push(child); commands.push({ label, command, args, cwd: root, pid: child.pid });
    let output = '';
    child.stdout.on('data', value => { output += value; }); child.stderr.on('data', value => { output += value; });
    child.on('error', error => { output += '\nProcess spawn failed: ' + error.message; });
    child.on('close', () => fs.writeFile(path.join(artifactDir, label + '.log'), output));
    return child;
  }
  const receipt = { version: 1, profile: 'generic', root, files, binarySHA256: sha(await fs.readFile(binary)), commands,
    connection: { clickHouse: 'explicit existing lab connection; credentials omitted', packages: environment.PACKAGES_URL }, cleanup: false };
  async function stop() {
    for (const child of children) if (child.pid && child.exitCode === null) child.kill('SIGTERM');
    await Promise.all(children.map(child => !child.pid || child.exitCode !== null || child.signalCode !== null ? Promise.resolve() : new Promise(resolve => {
      const timeout = setTimeout(() => { child.kill('SIGKILL'); }, 5000);
      child.once('close', () => { clearTimeout(timeout); resolve(); });
    })));
    receipt.cleanup = children.every(child => !child.pid || child.exitCode !== null || child.signalCode !== null);
    receipt.captureFiles = await inventory(path.join(root, 'captures'));
    await fs.cp(path.join(root, 'captures'), path.join(artifactDir, 'capture-store'), { recursive: true });
    await fs.writeFile(path.join(artifactDir, 'runtime.json'), JSON.stringify(receipt, null, 2));
    await fs.rm(root, { recursive: true, force: true });
  }
  try {
    await launch(binary, ['serve-http', '--profile', 'generic', '--listen', '127.0.0.1:18089', '--investigations-dir', path.join(root, 'captures')], 'go');
    await launch(process.env.PYTHON || 'python3', ['server.py'], 'ui');
    for (const url of ['http://127.0.0.1:18089/capabilities', base + '/api/ui-config']) {
      let ready = false;
      for (let attempt = 0; attempt < 100; attempt++) {
        if (children.some(child => child.exitCode !== null)) throw Error('Generic startup exited before readiness');
        try { const response = await fetch(url); if (response.ok) { ready = true; break; } } catch {}
        await new Promise(resolve => setTimeout(resolve, 100));
      }
      if (!ready) throw Error('Generic runtime readiness failed: ' + url);
    }
    await fs.writeFile(path.join(artifactDir, 'runtime.json'), JSON.stringify(receipt, null, 2));
    return { base, root, receipt, environment, stop };
  } catch (error) { await stop(); throw error; }
}
module.exports = { startRuntime, inventory, base };
