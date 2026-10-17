import unittest
from unittest.mock import Mock
from queue_telemetry import WorkerQueue, wait_ns


class QueueTests(unittest.TestCase):
    def test_active_work_is_bounded_and_released_on_exception(self):
        q = WorkerQueue(Mock(), capacity=1)
        with self.assertRaises(RuntimeError):
            with q.working():
                self.assertEqual(q.worker_value('utilization'), 1)
                with q.working():
                    pass
        self.assertEqual(q.worker_value('active'), 0)
        self.assertEqual(q.worker_value('utilization'), 0)

    def test_broker_failure_unknown_lag_and_pending_are_not_healthy_zero(self):
        client = Mock()
        q = WorkerQueue(client)
        client.xinfo_groups.return_value = [{'name': 'inventory', 'lag': None, 'pending': 3}]
        self.assertIsNone(q.broker_value('lag'))
        self.assertEqual(q.broker_value('pending'), 3)
        client.xinfo_groups.side_effect = OSError('unavailable')
        self.assertIsNone(q.broker_value('pending'))

    def test_wait_uses_stream_enqueue_time_and_rejects_clock_violation(self):
        self.assertEqual(wait_ns('1700000000000-5', 1700000000000000001), 1)
        self.assertEqual(wait_ns('0-0', 0), 0)
        self.assertIsNone(wait_ns('1700000000000-5', 1699999999999999999))
