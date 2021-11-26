from . import AgentCheck


class HelloPython2(AgentCheck):
    def check(self, instance):
        print ">" * 20
        print "  hello_python_plugin2   >. < ohhhhhhhhhhhh success!!!!!!!"
        print "<" * 20


def test():
    print "  hello_python_plugin2 test function   >. <"
