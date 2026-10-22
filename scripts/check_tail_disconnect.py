"""Loopback lab check: aborted downstream SSE readers release upstream slots.

Run with no other live monitors. Does not write telemetry or reset data.
"""
import json
import socket
import time
from pathlib import Path

BODY = json.dumps({'environment': 'tail-disconnect-verification'}).encode()
REQUEST = (b'POST /api/tail HTTP/1.1\r\nHost: 127.0.0.1:8088\r\n'
           b'Content-Type: application/json\r\nConnection: close\r\nContent-Length: '
           + str(len(BODY)).encode() + b'\r\n\r\n' + BODY)


def connect():
    client = socket.create_connection(('127.0.0.1', 8088), timeout=10)
    try:
        client.sendall(REQUEST)
        header = b''
        while b'\r\n\r\n' not in header:
            block = client.recv(4096)
            if not block:
                raise AssertionError('SSE closed before headers')
            header += block
            if len(header) > 65536:
                raise AssertionError('Oversized response headers')
        return client, int(header.split(b' ', 2)[1])
    except Exception:
        client.close()
        raise


def main():
    statuses = []
    for _ in range(8):
        client, status = connect()
        client.close()
        statuses.append(status)
        assert status == 200, ('aborted subscriptions occupied upstream slots', statuses)
    # Cancellation is concurrent with source completion: retry only the admission
    # probe, within one second, before checking the unchanged four-slot capacity.
    deadline = time.monotonic() + 1
    while True:
        held = []
        try:
            for _ in range(4):
                client, status = connect()
                held.append(client)
                if status != 200:
                    raise AssertionError(('slot admission', status))
            extra, status = connect()
            extra.close()
            assert status == 429, ('capacity changed', status)
            break
        except AssertionError:
            if time.monotonic() >= deadline:
                raise
            time.sleep(0.05)
        finally:
            for client in held:
                client.close()
    result = {'rapidAbortStatuses': statuses, 'simultaneousSlots': 4, 'fifthStatus': 429}
    Path('artifacts/stream-entry-proxy-cancel.json').write_text(json.dumps(result, indent=2))
    print(json.dumps(result))


if __name__ == '__main__':
    main()
