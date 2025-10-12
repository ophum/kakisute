#from prometheus_client import CollectorRegistry, multiprocess
#
#def when_ready(_):
#    multiprocess.MultiProcessCollector(CollectorRegistry())
#
#def child_exit(_, worker):
#    multiprocess.mark_process_dead(worker.pid)

wsgi_app = 'prom_test.wsgi:application'

daemon = False

bind = '0.0.0.0:8000'

workers = 2


