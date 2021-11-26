import time

from . import AgentCheck


class HelloPython(AgentCheck):
    def check(self, instance):
        print ">" * 20
        print "  hello_python_plugin   >. < ohhhhhhhhhhhh success!!!!!!!"
        print "<" * 20


def test():
    time.sleep(3)
    print "  hello_python_plugin test function   >. <"


if __name__ == '__main__':
    test()
