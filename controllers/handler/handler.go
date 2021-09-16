package handler

import (
	"context"
	"net/http"
	"strings"

	utilyaml "github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/horis233/k8s-deprecation-checker/controllers/utils"
)

type Recorder struct {
	Client     client.Client
	Namespaces []string
	decoder    *admission.Decoder
}

type DeprecatedObjectList struct {
	Group   string             `json:"group"`
	Version string             `json:"version"`
	Kind    string             `json:"kind"`
	Objects []DeprecatedObject `json:"objects"`
}
type DeprecatedObject struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// Handle will record deprecated resources
func (r *Recorder) Handle(ctx context.Context, req admission.Request) admission.Response {
	klog.Infof("Webhook is invoked by resource %s/%s, created by %s", req.AdmissionRequest.Namespace, req.AdmissionRequest.Name, req.UserInfo.Username)

	if len(r.Namespaces) != 0 {
		requesterNs := strings.Split(req.UserInfo.Username, ":")[2]
		requesterName := strings.Split(req.UserInfo.Username, ":")[3]
		var find bool
		// r.Namespaces = append(r.Namespaces, "openshift-operator-lifecycle-manager")
		for _, ns := range r.Namespaces {
			if requesterNs == ns {
				find = true
				break
			}
		}
		if !find {
			klog.Infof("Requester %s/%s is filtered", requesterNs, requesterName)
			return admission.Allowed("")
		}
	}

	var obj DeprecatedObject
	if req.Namespace == "" {
		obj = DeprecatedObject{
			Name: req.Name,
		}
	} else {
		obj = DeprecatedObject{
			Name:      req.Name,
			Namespace: req.Namespace,
		}
	}

	apiFromRequest := DeprecatedObjectList{
		Group:   req.Kind.Group,
		Version: req.Kind.Version,
		Kind:    req.Kind.Kind,
		Objects: []DeprecatedObject{
			obj,
		},
	}

	err := UpdateConfigmap(ctx, r.Client, apiFromRequest)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	return admission.Allowed("")
}

func AddtoReport(apiReport []DeprecatedObjectList, pendingApi DeprecatedObjectList) []DeprecatedObjectList {
	apiMap := make(map[string]int)
	apiObjMap := make(map[string][]DeprecatedObject)
	for i, objList := range apiReport {
		apiMap[objList.Group+objList.Version+objList.Kind] = i
		for _, obj := range objList.Objects {
			apiObjMap[obj.Name+obj.Namespace+objList.Group+objList.Version+objList.Kind] = objList.Objects
		}
	}
	if objIndex, apiFound := apiMap[pendingApi.Group+pendingApi.Version+pendingApi.Kind]; apiFound {
		_, objFound := apiObjMap[pendingApi.Objects[0].Name+pendingApi.Objects[0].Namespace+pendingApi.Group+pendingApi.Version+pendingApi.Kind]
		if objFound {
			return apiReport
		}
		apiReport[objIndex].Objects = append(apiReport[objIndex].Objects, pendingApi.Objects[0])
		return apiReport
	}
	apiReport = append(apiReport, pendingApi)
	return apiReport
}

func UpdateConfigmap(ctx context.Context, client client.Client, apiFromRequest DeprecatedObjectList) error {
	cm := &corev1.ConfigMap{}
	ns, err := utils.GetOperatorNamespace()
	if err != nil {
		klog.Error(err)
		return err
	}
	var apiSlice []DeprecatedObjectList

	err = client.Get(ctx, types.NamespacedName{Namespace: ns, Name: "deprecated-api-report"}, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			cm.SetName("deprecated-api-report")
			cm.SetNamespace(ns)
			apiSlice = AddtoReport(apiSlice, apiFromRequest)
			rawData, err := utilyaml.Marshal(apiSlice)
			if err != nil {
				klog.Error(err)
				return err
			}
			cm.Data = make(map[string]string)
			cm.Data["deprecated-api-report.yaml"] = string(rawData)
			err = client.Create(ctx, cm)
			if err != nil {
				klog.Error(err)
				return err
			}
		} else {
			klog.Error(err)
			return err
		}
	}
	deprecatedApiReport := cm.Data["deprecated-api-report.yaml"]
	if err := utilyaml.Unmarshal([]byte(deprecatedApiReport), &apiSlice); err != nil {
		klog.Error(err)
		return err
	}
	apiSlice = AddtoReport(apiSlice, apiFromRequest)
	rawData, err := utilyaml.Marshal(apiSlice)
	if err != nil {
		klog.Error(err)
		return err
	}
	cm.Data["deprecated-api-report.yaml"] = string(rawData)
	err = client.Update(ctx, cm)
	return err
}
