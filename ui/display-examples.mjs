/* Embedding reference only. These supplied fixtures never acquire source data. */
const summary = JSON.parse(document.getElementById('display408-summary-data').textContent);
const groups = JSON.parse(document.getElementById('display408-ranked-data').textContent);
const kind = document.getElementById('display408-kind');
const status = document.getElementById('display408-status');

// Own properties set before registration are adopted when the element upgrades.
const ranked = document.createElement('goo-chart');
ranked.id = 'display408-ranked';
ranked.snapshot = groups;
ranked.display = {kind: 'bar', category: 'category', value: 'count', title: 'Supplied groups'};
ranked.textContent = 'api: 12 records; worker: 8 records; archive: 0 records.';
document.getElementById('display408-ranked-host').append(ranked);

await import('./display.mjs');

// Detached property assignment uses the same supplied summary as the HTML above.
const table = document.createElement('goo-chart');
table.id = 'display408-summary-table';
table.snapshot = summary;
table.display = {kind: 'table', title: 'Supplied summary as a table'};
document.getElementById('display408-summary-table-host').append(table);

kind.disabled = false;
kind.addEventListener('change', () => {
  ranked.display = {...ranked.display, kind: kind.value};
  status.textContent = `${kind.selectedOptions[0].textContent} view · same supplied groups.`;
});
ranked.addEventListener('goo-select', event => {
  const row = groups.rows.find(row => row.key === event.detail?.key);
  if (row) status.textContent = `Selected ${row.values.category} · ${row.values.count} supplied records.`;
});
status.textContent = 'Supplied examples ready.';
