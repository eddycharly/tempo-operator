package webhooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	configv1alpha1 "github.com/grafana/tempo-operator/apis/config/v1alpha1"
	"github.com/grafana/tempo-operator/apis/tempo/v1alpha1"
)

func TestMonolithicValidate(t *testing.T) {
	tests := []struct {
		name       string
		ctrlConfig configv1alpha1.ProjectConfig
		tempo      v1alpha1.TempoMonolithic
		warnings   admission.Warnings
		errors     field.ErrorList
	}{
		{
			name: "valid instance",
			tempo: v1alpha1.TempoMonolithic{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sample",
					Namespace: "default",
				},
				Spec: v1alpha1.TempoMonolithicSpec{},
			},
			warnings: admission.Warnings{},
			errors:   field.ErrorList{},
		},

		// Jaeger UI
		{
			name: "JaegerUI ingress enabled but Jaeger UI disabled",
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					JaegerUI: &v1alpha1.MonolithicJaegerUISpec{
						Ingress: &v1alpha1.MonolithicJaegerUIIngressSpec{
							Enabled: true,
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors: field.ErrorList{field.Invalid(
				field.NewPath("spec", "jaegerui", "ingress", "enabled"),
				true,
				"Jaeger UI must be enabled to create an ingress for Jaeger UI",
			)},
		},
		{
			name: "JaegerUI route enabled but Jaeger UI disabled",
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					JaegerUI: &v1alpha1.MonolithicJaegerUISpec{
						Route: &v1alpha1.MonolithicJaegerUIRouteSpec{
							Enabled: true,
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors: field.ErrorList{field.Invalid(
				field.NewPath("spec", "jaegerui", "route", "enabled"),
				true,
				"Jaeger UI must be enabled to create a route for Jaeger UI",
			)},
		},
		{
			name: "JaegerUI route enabled but openShiftRoute feature gate not set",
			ctrlConfig: configv1alpha1.ProjectConfig{
				Gates: configv1alpha1.FeatureGates{
					OpenShift: configv1alpha1.OpenShiftFeatureGates{
						OpenShiftRoute: false,
					},
				},
			},
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					JaegerUI: &v1alpha1.MonolithicJaegerUISpec{
						Enabled: true,
						Route: &v1alpha1.MonolithicJaegerUIRouteSpec{
							Enabled: true,
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors: field.ErrorList{field.Invalid(
				field.NewPath("spec", "jaegerui", "route", "enabled"),
				true,
				"the openshiftRoute feature gate must be enabled to create a route for Jaeger UI",
			)},
		},
		{
			name: "JaegerUI route enabled and openshiftRoute feature gate set",
			ctrlConfig: configv1alpha1.ProjectConfig{
				Gates: configv1alpha1.FeatureGates{
					OpenShift: configv1alpha1.OpenShiftFeatureGates{
						OpenShiftRoute: true,
					},
				},
			},
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					JaegerUI: &v1alpha1.MonolithicJaegerUISpec{
						Enabled: true,
						Route: &v1alpha1.MonolithicJaegerUIRouteSpec{
							Enabled: true,
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors:   field.ErrorList{},
		},

		// observability
		{
			name: "serviceMonitors enabled but prometheusOperator feature gate not set",
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					Observability: &v1alpha1.MonolithicObservabilitySpec{
						Metrics: &v1alpha1.MonolithicObservabilityMetricsSpec{
							ServiceMonitors: &v1alpha1.MonolithicObservabilityMetricsServiceMonitorsSpec{
								Enabled: true,
							},
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors: field.ErrorList{field.Invalid(
				field.NewPath("spec", "observability", "metrics", "serviceMonitors", "enabled"),
				true,
				"the prometheusOperator feature gate must be enabled to create ServiceMonitors for Tempo components",
			)},
		},
		{
			name: "prometheusRules enabled but prometheusOperator feature gate not set",
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					Observability: &v1alpha1.MonolithicObservabilitySpec{
						Metrics: &v1alpha1.MonolithicObservabilityMetricsSpec{
							PrometheusRules: &v1alpha1.MonolithicObservabilityMetricsPrometheusRulesSpec{
								Enabled: true,
							},
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors: field.ErrorList{field.Invalid(
				field.NewPath("spec", "observability", "metrics", "prometheusRules", "enabled"),
				true,
				"the prometheusOperator feature gate must be enabled to create PrometheusRules for Tempo components",
			)},
		},
		{
			name: "prometheusRules enabled but serviceMonitors disabled",
			ctrlConfig: configv1alpha1.ProjectConfig{
				Gates: configv1alpha1.FeatureGates{
					PrometheusOperator: true,
				},
			},
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					Observability: &v1alpha1.MonolithicObservabilitySpec{
						Metrics: &v1alpha1.MonolithicObservabilityMetricsSpec{
							PrometheusRules: &v1alpha1.MonolithicObservabilityMetricsPrometheusRulesSpec{
								Enabled: true,
							},
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors: field.ErrorList{field.Invalid(
				field.NewPath("spec", "observability", "metrics", "prometheusRules", "enabled"),
				true,
				"serviceMonitors must be enabled to create PrometheusRules (the rules alert based on collected metrics)",
			)},
		},
		{
			name: "dataSource enabled but grafanaOperator feature gate not set",
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					Observability: &v1alpha1.MonolithicObservabilitySpec{
						Grafana: &v1alpha1.MonolithicObservabilityGrafanaSpec{
							DataSource: &v1alpha1.MonolithicObservabilityGrafanaDataSourceSpec{
								Enabled: true,
							},
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors: field.ErrorList{field.Invalid(
				field.NewPath("spec", "observability", "grafana", "dataSource", "enabled"),
				true,
				"the grafanaOperator feature gate must be enabled to create a data source for Tempo",
			)},
		},
		{
			name: "valid observability config",
			ctrlConfig: configv1alpha1.ProjectConfig{
				Gates: configv1alpha1.FeatureGates{
					PrometheusOperator: true,
					GrafanaOperator:    true,
				},
			},
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					Observability: &v1alpha1.MonolithicObservabilitySpec{
						Metrics: &v1alpha1.MonolithicObservabilityMetricsSpec{
							ServiceMonitors: &v1alpha1.MonolithicObservabilityMetricsServiceMonitorsSpec{
								Enabled: true,
							},
							PrometheusRules: &v1alpha1.MonolithicObservabilityMetricsPrometheusRulesSpec{
								Enabled: true,
							},
						},
						Grafana: &v1alpha1.MonolithicObservabilityGrafanaSpec{
							DataSource: &v1alpha1.MonolithicObservabilityGrafanaDataSourceSpec{
								Enabled: true,
							},
						},
					},
				},
			},
			warnings: admission.Warnings{},
			errors:   field.ErrorList{},
		},

		// extra config
		{
			name: "extra config warning",
			tempo: v1alpha1.TempoMonolithic{
				Spec: v1alpha1.TempoMonolithicSpec{
					ExtraConfig: &v1alpha1.ExtraConfigSpec{
						Tempo: apiextensionsv1.JSON{Raw: []byte(`{}`)},
					},
				},
			},
			warnings: admission.Warnings{"overriding Tempo configuration could potentially break the deployment, use it carefully"},
			errors:   field.ErrorList{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &k8sFake{}
			v := &monolithicValidator{
				client:     client,
				ctrlConfig: test.ctrlConfig,
			}

			warnings, errors := v.validateTempoMonolithic(context.Background(), test.tempo)
			require.Equal(t, test.warnings, warnings)
			require.Equal(t, test.errors, errors)
		})
	}
}
