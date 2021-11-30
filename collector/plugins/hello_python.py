import time

from . import AgentCheck


class HelloPython(AgentCheck):
    def check(self, instance):
        print "  hello_python_plugin >. < success!", instance
        self.gauge("demo_gauge0", 321, tags={"tagk:1123", "tagv"}, hostname=None, device_name=None)
        self.gauge("demo_gauge10", 123, tags={"tagkk", "tagvv"}, hostname=None, device_name=None)


