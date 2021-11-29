from . import AgentCheck


class HelloPython2(AgentCheck):
    def check(self, instance):
        print "  hello_python_plugin2   >. < ohhhhhhhhhhhh success!!!!!!!", instance


def test():
    print "  hello_python_plugin2 test function   >. <"
