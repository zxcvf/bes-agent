package py

/*
#cgo pkg-config: python-2.7
#cgo windows LDFLAGS: -Wl,--allow-multiple-definition
#include "stdlib.h"
#include <Python.h>


typedef enum {
    MT_FIRST = 0,
    GAUGE = MT_FIRST,
    RATE,
    COUNT,
    MONOTONIC_COUNT,
    COUNTER,
    HISTOGRAM,
    HISTORATE,
    MT_LAST = HISTORATE
} MetricType;

PyObject* SubmitMetric(PyObject*, char*, MetricType, char*, float, PyObject*, char*);
PyObject* SubmitServiceCheck(PyObject*, char*, char*, int, PyObject*, char*, char*);
PyObject* SubmitEvent(PyObject*, char*, PyObject*);

//_must_ be in the same order as the MetricType enum
// multiple definition when using cgo -> __attribute__((weak))
__attribute__((weak))
char* MetricTypeNames[] = {
    "GAUGE",
    "RATE",
    "COUNT",
    "MONOTONIC_COUNT",
    "COUNTER",
    "HISTOGRAM",
    "HISTORATE"
};

static PyObject *submit_metric(PyObject *self, PyObject *args) {
    PyObject *check = NULL;
    int mt;
    char *name;
    float value;
    PyObject *tags = NULL;
    char *hostname;
    char *check_id;

    PyGILState_STATE gstate;
    gstate = PyGILState_Ensure();

    // aggregator.submit_metric(self, check_id, aggregator.metric_type.GAUGE, name, value, tags, hostname)
    if (!PyArg_ParseTuple(args, "OsisfOs", &check, &check_id, &mt, &name, &value, &tags, &hostname)) {
      PyGILState_Release(gstate);
      return NULL;
    }

    PyGILState_Release(gstate);
    return SubmitMetric(check, check_id, mt, name, value, tags, hostname);
}

static PyMethodDef AggMethods[] = {
  {"submit_metric", (PyCFunction)submit_metric, METH_VARARGS, "Submit metrics to the aggregator."},
  {NULL, NULL}  // guards
};

__attribute__((weak))
PyObject* _none() {
	Py_RETURN_NONE;
}

__attribute__((weak))
int _is_none(PyObject *o) {
  return o == Py_None;
}

__attribute__((weak))
void initaggregator()
{
  printf("CGO: init aggregator in C \n");
  PyGILState_STATE gstate;
  gstate = PyGILState_Ensure();

  PyObject *m = Py_InitModule("aggregator", AggMethods);

  int i;
  for (i=MT_FIRST; i<=MT_LAST; i++) {
    PyModule_AddIntConstant(m, MetricTypeNames[i], i);
  }

  PyGILState_Release(gstate);
}

__attribute__((weak))
int _PyDict_Check(PyObject *o) {
  return PyDict_Check(o);
}

__attribute__((weak))
int _PyInt_Check(PyObject *o) {
  return PyInt_Check(o);
}

__attribute__((weak))
int _PyString_Check(PyObject *o) {
  return PyString_Check(o);
}

__attribute__((weak))
PyObject* PySequence_Fast_Get_Item(PyObject *o, Py_ssize_t i)
{
  return PySequence_Fast_GET_ITEM(o, i);
}

__attribute__((weak))
Py_ssize_t PySequence_Fast_Get_Size(PyObject *o)
{
  return PySequence_Fast_GET_SIZE(o);
}

*/
import "C"

// 在此export到头部C对SubmitMetric的声明中
// https://stackoverflow.com/questions/58606884/multiple-definition-when-using-cgo
// 参照 https://pkg.go.dev/cmd/cgo#hdr-C_references_to_Go
//export SubmitMetric
func SubmitMetric(check *C.PyObject, checkID *C.char, mt C.MetricType, name *C.char, value C.float, tags *C.PyObject, hostname *C.char) *C.PyObject {
	return C._none()
}

func InitCAggregator() {
	C.initaggregator()
}
