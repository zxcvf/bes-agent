import time

from . import AgentCheck


class HelloPython(AgentCheck):
    def check(self, instance):
        print "  hello_python_plugin   >. < ohhhhhhhhhhhh success!!!!!!!", instance
        self.gauge("demo_gauge", 321, tags={"tagk": "tagv"}, hostname=None, device_name=None)
        self.gauge("demo_gauge1", 123, tags={"tagkk": "tagvv"}, hostname=None, device_name=None)


def test():
    time.sleep(3)
    print "  hello_python_plugin test function   >. <"


if __name__ == '__main__':
    test()
