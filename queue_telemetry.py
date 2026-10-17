"""Redis Streams adapter for the demo's custom worker/queue metric contract.

OTel messaging attributes describe identity; lab.* metric names are deliberately
custom. A failed observation emits no gauge points rather than a healthy zero.
"""
import contextlib
import threading
import time
from opentelemetry.metrics import Observation

CAPACITY = 7  # actual dispatch bound, not the executor's theoretical default
ATTRS = {'messaging.system': 'redis', 'messaging.destination.name': 'reservations',
         'messaging.consumer.group.name': 'inventory'}


class WorkerQueue:
    def __init__(self, client, capacity=CAPACITY):
        if capacity <= 0:
            raise ValueError('capacity must be positive')
        self.client, self.capacity = client, capacity
        self.active = 0
        self.lock = threading.Lock()

    @contextlib.contextmanager
    def working(self):
        with self.lock:
            if self.active >= self.capacity:
                raise RuntimeError('worker capacity exceeded')
            self.active += 1
        try:
            yield
        finally:
            with self.lock:
                self.active -= 1

    def worker_value(self, metric):
        with self.lock:
            return {'active': self.active, 'capacity': self.capacity,
                    'utilization': self.active / self.capacity}[metric]

    def broker_value(self, key):
        try:
            group = next(g for g in self.client.xinfo_groups('reservations') if g['name'] == 'inventory')
            value = group.get(key)
            # Redis lag may be unknown after particular stream operations.
            return value if type(value) is int and value >= 0 else None
        except Exception:
            return None

    def register(self, meter):
        for metric, unit, getter in [
            ('lab.queue.depth', '{message}', lambda: self.broker_value('lag')),
            ('lab.queue.pending', '{message}', lambda: self.broker_value('pending')),
            ('lab.worker.active', '{worker}', lambda: self.worker_value('active')),
            ('lab.worker.capacity', '{worker}', lambda: self.worker_value('capacity')),
            ('lab.worker.utilization', '1', lambda: self.worker_value('utilization')),
        ]:
            def observe(options, read=getter):
                value = read()
                return [] if value is None else [Observation(value, ATTRS)]
            meter.create_observable_gauge(metric, callbacks=[observe], unit=unit)


def wait_ns(message_id, started_ns=None):
    """Redis-assigned enqueue ms to first worker start; None on clock violation."""
    enqueue_ns = int(message_id.split('-')[0]) * 1_000_000
    elapsed = (time.time_ns() if started_ns is None else started_ns) - enqueue_ns
    return elapsed if elapsed >= 0 else None
