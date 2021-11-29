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


static PyObject *submit_service_check(PyObject *self, PyObject *args) {
    PyObject *check = NULL;
    char *name;
    int status;
    PyObject *tags = NULL;
    char *hostname;
    char *message = NULL;
    char *check_id;

    PyGILState_STATE gstate;
    gstate = PyGILState_Ensure();

    // aggregator.submit_service_check(self, check_id, name, status, tags, hostname, message)
    if (!PyArg_ParseTuple(args, "OssiOss", &check, &check_id, &name, &status, &tags, &hostname, &message)) {
      PyGILState_Release(gstate);
      return NULL;
    }

    PyGILState_Release(gstate);
    return SubmitServiceCheck(check, check_id, name, status, tags, hostname, message);
}


static PyObject *submit_event(PyObject *self, PyObject *args) {
    PyObject *check = NULL;
    PyObject *event = NULL;
    char *check_id;

    PyGILState_STATE gstate;
    gstate = PyGILState_Ensure();

    // aggregator.submit_event(self, check_id, event)
    if (!PyArg_ParseTuple(args, "OsO", &check, &check_id, &event)) {
      PyGILState_Release(gstate);
      return NULL;
    }

    PyGILState_Release(gstate);
    return SubmitEvent(check, check_id, event);
}


static PyMethodDef AggMethods[] = {
  {"submit_metric", (PyCFunction)submit_metric, METH_VARARGS, "Submit metrics to the aggregator."},
  {"submit_service_check", (PyCFunction)submit_service_check, METH_VARARGS, "Submit service checks to the aggregator."},
  {"submit_event", (PyCFunction)submit_event, METH_VARARGS, "Submit events to the aggregator."},
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
import (
	"errors"
	"fmt"
	"unsafe"
)

// 在此export到头部C对SubmitMetric的声明中
// https://stackoverflow.com/questions/58606884/multiple-definition-when-using-cgo
// 参照 https://pkg.go.dev/cmd/cgo#hdr-C_references_to_Go
//export SubmitMetric
func SubmitMetric(
	check *C.PyObject,
	checkID *C.char,
	mt C.MetricType,
	name *C.char,
	value C.float,
	tags *C.PyObject,
	hostname *C.char,
) *C.PyObject {

	goCheckID := C.GoString(checkID)
	_name := C.GoString(name)
	_value := float64(value)
	_tags, err := extractTags(tags)
	_hostname := C.GoString(hostname)
	fmt.Println("> SubmitMetric", goCheckID, _name, _value, _tags, err, _hostname)
	//fmt.Println("> PythonAggregatorPool", PythonAggregatorPool)

	agg := PythonAggregatorPool[ID(goCheckID)]

	fmt.Println(agg)

	// todo 使用AGG 上报
	fmt.Println("goCheckID", goCheckID)

	return nil
	//goCheckID := C.GoString(checkID)
	//var sender aggregator.Sender
	//var err error
	//
	//sender, err = aggregator.GetSender(chk.ID(goCheckID))
	//
	//if err != nil || sender == nil {
	//	log.Errorf("Error submitting metric to the Sender: %v", err)
	//	return C._none()
	//}
	//
	//_name := C.GoString(name)
	//_value := float64(value)
	//_tags, err := extractTags(tags)
	//if err != nil {
	//	log.Error(err)
	//	return nil
	//}
	//_hostname := C.GoString(hostname)
	//
	//switch mt {
	//case C.GAUGE:
	//	sender.Gauge(_name, _value, _hostname, _tags)
	//case C.RATE:
	//	sender.Rate(_name, _value, _hostname, _tags)
	//case C.COUNT:
	//	sender.Count(_name, _value, _hostname, _tags)
	//case C.MONOTONIC_COUNT:
	//	sender.MonotonicCount(_name, _value, _hostname, _tags)
	//case C.COUNTER:
	//	sender.Counter(_name, _value, _hostname, _tags)
	//case C.HISTOGRAM:
	//	sender.Histogram(_name, _value, _hostname, _tags)
	//case C.HISTORATE:
	//	sender.Historate(_name, _value, _hostname, _tags)
	//}
	//
	//return C._none()
}

//export SubmitServiceCheck
func SubmitServiceCheck(check *C.PyObject, checkID *C.char, name *C.char, status C.int, tags *C.PyObject, hostname *C.char, message *C.char) *C.PyObject {
	return C._none()
}

//export SubmitEvent
func SubmitEvent(check *C.PyObject, checkID *C.char, event *C.PyObject) *C.PyObject {

	return C._none()
}

// extractEventFromDict returns an `Event` populated with the fields of the passed event py object
// The caller needs to check the returned `error`, any non-nil value indicates that the error flag is set
// on the python interpreter.
//func extractEventFromDict(event *C.PyObject) (metrics.Event, error) {
//	return
//}

// extractTags returns a slice with the contents of the passed non-nil py object.
// The caller needs to check the returned `error`, any non-nil value indicates that the error flag is set
// on the python interpreter.
func extractTags(tags *C.PyObject) (_tags []string, err error) {
	if !isNone(tags) {
		if int(C.PySequence_Check(tags)) == 0 {
			fmt.Println("Submitted `tags` is not a sequence, ignoring tags")
			return
		}

		errMsg := C.CString("expected tags to be a sequence")
		defer C.free(unsafe.Pointer(errMsg))

		var seq *C.PyObject
		seq = C.PySequence_Fast(tags, errMsg) // seq is a new reference, has to be decref'd
		if seq == nil {
			err = errors.New("can't iterate on tags")
			return
		}
		defer C.Py_DecRef(seq)

		var i C.Py_ssize_t
		for i = 0; i < C.PySequence_Fast_Get_Size(seq); i++ {
			item := C.PySequence_Fast_Get_Item(seq, i) // `item` is borrowed, no need to decref
			if int(C._PyString_Check(item)) == 0 {
				fmt.Println("One of the submitted tag is not a string, ignoring it")
				continue
			}
			// at this point we're sure that `item` is a string, no further error checking needed
			_tags = append(_tags, C.GoString(C.PyString_AsString(item)))
		}
	}

	return
}

func isNone(o *C.PyObject) bool {
	return int(C._is_none(o)) != 0
}

func InitCAggregator() {
	C.initaggregator()
}
