"""Discover the explicitly adopted Go route-registration forms for API tooling.

This is source indexing, not a general Go parser or runtime route registry.
Semantic policy remains in Go and is checked against the published contract.
"""
import re


def route_sources(root):
    sources = {}
    files = sorted((root / 'cmd/goo').glob('*http.go'))
    files.extend(root / path for path in (
        'cmd/goo/serve_ui.go',
        'internal/recordhttp/routes.go',
        'internal/nativehttp/routes.go',
        'internal/explanationhttp/routes.go',
        'internal/capturehttp/routes.go',
        'internal/capturehttp/live.go',
    ))
    pattern = re.compile(r'HandleFunc\("(GET|POST) ([^" ]+)"|register(?:Read|Plan)\(mux,\s*"([^"]+)"')
    for file in files:
        text = file.read_text()
        matches = list(pattern.finditer(text))
        for index, match in enumerate(matches):
            method, path = ('post', match[3]) if match[3] else (match[1].lower(), match[2])
            key = method, path
            if key in sources:
                raise ValueError(f'duplicate Go route: {method} {path}')
            end = matches[index+1].start() if index+1 < len(matches) else len(text)
            sources[key] = str(file.relative_to(root)), text[match.end():end]
    sources['post', '/queries/{id}/cancel'] = 'internal/queryhttp/budget.go', ''
    return sources
