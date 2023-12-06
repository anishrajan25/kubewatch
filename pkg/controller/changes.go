package controller

import (
	"encoding/json"
	"reflect"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type IChanges interface {
	GetInitialValues() map[string]interface{}
	GetCurrentValues() map[string]interface{}
	String() string
}

type changes struct {
	InitialValues map[string]interface{} `json:"initial_values"`
	CurrentValues map[string]interface{} `json:"current_values"`
}

func (c *changes) GetInitialValues() map[string]interface{} {
	return c.InitialValues
}

func (c *changes) GetCurrentValues() map[string]interface{} {
	return c.CurrentValues
}

func (c *changes) String() string {
	b, e := json.Marshal(c)
	if e != nil {
		return e.Error()
	}

	return string(b)
}

func GetChanges(oldObj, newObj runtime.Object, compareFields []*config.CompareField) IChanges {
	oldObject := getUnstructuredObject(oldObj)
	newObject := getUnstructuredObject(newObj)

	if oldObject == nil || newObject == nil {
		logrus.Errorf("Either the new object or old object is not a valid type to evaluate changes.")
		return nil
	}

	return &changes{
		InitialValues: getFieldValues(oldObject, compareFields),
		CurrentValues: getFieldValues(newObject, compareFields),
	}
}

func getUnstructuredObject(obj runtime.Object) *unstructured.Unstructured {
	object, ok := obj.(*unstructured.Unstructured)
	if !ok {
		logrus.Errorf("Invalid object type - %s provided for evaluating changes.", reflect.TypeOf(obj))
		return nil
	}

	return object
}

func getFieldValues(object *unstructured.Unstructured, compareFields []*config.CompareField) map[string]interface{} {
	fieldValues := map[string]interface{}{}

	for _, field := range compareFields {
		fieldValues[field.Name] = getNestedString(object.Object, field.Path...)
	}

	return fieldValues
}

func getNestedString(obj map[string]interface{}, fields ...string) string {
	val, found, err := unstructured.NestedString(obj, fields...)
	if !found || err != nil {
		return ""
	}
	return val
}
