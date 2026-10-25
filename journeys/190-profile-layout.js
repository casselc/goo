const { expect } = require('@playwright/test');
async function profileLayout(page, mobile = false) {
  const value = await page.locator('.service-dashboard .dashboard-profiles').evaluate(nav => {
    const rect = element => { const r = element.getBoundingClientRect(); return { left: r.left, right: r.right, top: r.top, bottom: r.bottom, width: r.width }; };
    const style = getComputedStyle(nav);
    return { bounds: rect(nav), display: style.display, columns: style.gridTemplateColumns,
      buttons: [...nav.querySelectorAll('button[data-dashboard-profile]')].map(button => {
        const text = button.firstChild;
        const words = [...text.textContent.matchAll(/\S+/g)].map(match => {
          const range = document.createRange(); range.setStart(text, match.index); range.setEnd(text, match.index + match[0].length);
          return { text: match[0], rects: [...range.getClientRects()].map(r => ({ left: r.left, right: r.right, top: r.top, bottom: r.bottom })) };
        });
        return { key: button.dataset.dashboardProfile, text: button.textContent, selected: button.getAttribute('aria-pressed'), disabled: button.disabled,
          bounds: rect(button), scroll: button.scrollWidth, width: button.clientWidth, words };
      }) };
  });
  expect(value.buttons).toHaveLength(6);
  expect(value.buttons.filter(button => button.selected === 'true').map(button => button.key)).toEqual(['messaging-consumer']);
  for (const button of value.buttons) {
    expect(button.bounds.left).toBeGreaterThanOrEqual(value.bounds.left - 1);
    expect(button.bounds.right).toBeLessThanOrEqual(value.bounds.right + 1);
    expect(button.scroll).toBeLessThanOrEqual(button.width + 1);
    for (const word of button.words) {
      expect(word.rects, 'word must not split vertically: ' + word.text).toHaveLength(1);
      expect(word.rects[0].left).toBeGreaterThanOrEqual(button.bounds.left - 1);
      expect(word.rects[0].right).toBeLessThanOrEqual(button.bounds.right + 1);
      expect(word.rects[0].top).toBeGreaterThanOrEqual(button.bounds.top - 1);
      expect(word.rects[0].bottom).toBeLessThanOrEqual(button.bounds.bottom + 1);
    }
  }
  if (mobile) {
    expect(value.display).toBe('grid');
    expect(value.columns.trim().split(/\s+/)).toHaveLength(2);
    expect(new Set(value.buttons.map(button => Math.round(button.bounds.left))).size).toBe(2);
  }
  return value;
}
module.exports = { profileLayout };
