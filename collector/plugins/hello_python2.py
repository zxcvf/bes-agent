# try:
#     # first, try to import the base class from new versions of the Agent
#     from datadog_checks.base import AgentCheck
# except ImportError:
#     # if the above failed, the check is running in Agent version < 6.6.0
#     from checks import AgentCheck
#
# # content of the special variable __version__ will be shown in the Agent status page
# __version__ = "1.0.0"
#
#
# class Hello2Python(AgentCheck):
#     def check(self, instance):
#         print "  hello_python_plugin   >. <"
#         print("*" * 50)
#         print("*" * 50)
#         print("*" * 50)
#         print("*" * 50)

def test():
    print "  hello_python_plugin2 test function   >. <"
