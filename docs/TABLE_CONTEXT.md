# Reading wide tables

The shared `goo-table` keeps caller-authored native HTML, labels and controls.
When content exceeds the viewport, its edges indicate which direction contains
more columns. The cue disappears when the table fits. Focus the existing region
and use the browser's scroll keys, trackpad or scrollbar to reach later fields.
No extra toolbar or announcement is added.

Ordinary quantitative, grouped-interval and retained native tables opt into
`sticky-first-column` and `sticky-header`. The first column retains category,
group or UTC observation/interval context while later values are visible. The
identity cell wraps within a readable bounded width. Overflowing grouped Table
rows also retain a compact UTC start time with an exact interval tooltip, so
repeated groups remain distinguishable beside later counts. Literal case, units, support and
exact timestamp tooltips remain supplied by the existing renderer. Stacked
mobile tables retain their own row layout. Dense scope/grouped viewports have a
scoped height bound so headers remain visible while reading longer tables.

```html
<goo-table aria-label="Grouped intervals" density="compact"
           sticky-header sticky-first-column>
  <table><!-- caller-owned caption, scoped headers and rows --></table>
</goo-table>
```

`data-overflow-inline` is derived presentation state (`start`, `end`, `both`, or
absent). Callers should not persist it. Local resize, content and scroll observers
update it, disconnect on removal and reconnect without replacing content.
Server-authored HTML and detached component construction use the same boundary.
The component performs no reads, publishes no selection/query events, and owns
no sorting, paging or saved definitions. Source and evidence ownership are unchanged.

See [shared components](UI_COMPONENTS.md) and [validation](../VALIDATION.md).
