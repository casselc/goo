const base = 'http://127.0.0.1:8088';
async function moduleFixture(page) {
  await page.route(base+'/module-fixture', route => route.fulfill({
    contentType: 'text/html; charset=utf-8',
    body: '<!doctype html><html lang="en"><head><meta name="viewport" content="width=device-width, initial-scale=1"></head><body></body></html>',
  }));
  await page.goto(base+'/module-fixture');
}
module.exports = {moduleFixture};
