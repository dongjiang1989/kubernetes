/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package corev1

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestPodPriority tests PodPriority function.
func TestPodPriority(t *testing.T) {
	p := int32(20)
	tests := []struct {
		name             string
		pod              *v1.Pod
		expectedPriority int32
	}{
		{
			name: "no priority pod resolves to static default priority",
			pod: &v1.Pod{
				Spec: v1.PodSpec{Containers: []v1.Container{
					{Name: "container", Image: "image"}},
				},
			},
			expectedPriority: 0,
		},
		{
			name: "pod with priority resolves correctly",
			pod: &v1.Pod{
				Spec: v1.PodSpec{Containers: []v1.Container{
					{Name: "container", Image: "image"}},
					Priority: &p,
				},
			},
			expectedPriority: p,
		},
	}
	for _, test := range tests {
		if PodPriority(test.pod) != test.expectedPriority {
			t.Errorf("expected pod priority: %v, got %v", test.expectedPriority, PodPriority(test.pod))
		}

	}
}

func TestMatchNodeSelectorTerms(t *testing.T) {
	type args struct {
		nodeSelector *v1.NodeSelector
		node         *v1.Node
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil terms",
			args: args{
				nodeSelector: nil,
				node:         nil,
			},
			want: false,
		},
		{
			name: "node label matches matchExpressions terms",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label_1": "label_1_val"}}},
			},
			want: true,
		},
		{
			name: "node field matches matchFields terms",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1"}},
			},
			want: true,
		},
		{
			name: "invalid node field requirement",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1", "host_2"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1"}},
			},
			want: false,
		},
		{
			name: "fieldSelectorTerm with node labels",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "not_host_1", Labels: map[string]string{
					"metadata.name": "host_1",
				}}},
			},
			want: false,
		},
		{
			name: "labelSelectorTerm with node fields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1"}},
			},
			want: false,
		},
		{
			name: "labelSelectorTerm and fieldSelectorTerm was set, but only node fields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val"},
						}},
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1"}},
			},
			want: false,
		},
		{
			name: "labelSelectorTerm and fieldSelectorTerm was set, both node fields and labels (both matched)",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val"},
						}},
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1"},
						}},
					}},
				},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1",
					Labels: map[string]string{
						"label_1": "label_1_val",
					}}},
			},
			want: true,
		},
		{
			name: "labelSelectorTerm and fieldSelectorTerm was set, both node fields and labels (one mismatched)",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val"},
						}},
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1",
					Labels: map[string]string{
						"label_1": "label_1_val-failed",
					}}},
			},
			want: false,
		},
		{
			name: "multi-selector was set, both node fields and labels (one mismatched)",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val"},
						}},
					},
					{
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1",
					Labels: map[string]string{
						"label_1": "label_1_val-failed",
					}}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := MatchNodeSelectorTerms(tt.args.node, tt.args.nodeSelector); got != tt.want {
				t.Errorf("MatchNodeSelectorTermsORed() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMatchNodeSelectorTermsStateless ensures MatchNodeSelectorTerms()
// is invoked in a "stateless" manner, i.e. nodeSelector should NOT
// be deeply modified after invoking
func TestMatchNodeSelectorTermsStateless(t *testing.T) {
	type args struct {
		nodeSelector *v1.NodeSelector
		node         *v1.Node
	}

	tests := []struct {
		name string
		args args
		want *v1.NodeSelector
	}{
		{
			name: "nil terms",
			args: args{
				nodeSelector: nil,
				node:         nil,
			},
			want: nil,
		},
		{
			name: "nodeLabels: preordered matchExpressions and nil matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val", "label_2_val"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label_1": "label_1_val"}}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "label_1",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"label_1_val", "label_2_val"},
					}},
				},
			}},
		},
		{
			name: "nodeLabels: unordered matchExpressions and nil matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_2_val", "label_1_val"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"label_1": "label_1_val"}}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "label_1",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"label_2_val", "label_1_val"},
					}},
				},
			}},
		},
		{
			name: "nodeFields: nil matchExpressions and preordered matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1", "host_2"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1"}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchFields: []v1.NodeSelectorRequirement{{
						Key:      "metadata.name",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"host_1", "host_2"},
					}},
				},
			}},
		},
		{
			name: "nodeFields: nil matchExpressions and unordered matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_2", "host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1"}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchFields: []v1.NodeSelectorRequirement{{
						Key:      "metadata.name",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"host_2", "host_1"},
					}},
				},
			}},
		},
		{
			name: "nodeLabels and nodeFields: ordered matchExpressions and ordered matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val", "label_2_val"},
						}},
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1", "host_2"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1",
					Labels: map[string]string{
						"label_1": "label_1_val",
					}}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "label_1",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"label_1_val", "label_2_val"},
					}},
					MatchFields: []v1.NodeSelectorRequirement{{
						Key:      "metadata.name",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"host_1", "host_2"},
					}},
				},
			}},
		},
		{
			name: "nodeLabels and nodeFields: ordered matchExpressions and unordered matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_1_val", "label_2_val"},
						}},
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_2", "host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1",
					Labels: map[string]string{
						"label_1": "label_1_val",
					}}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "label_1",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"label_1_val", "label_2_val"},
					}},
					MatchFields: []v1.NodeSelectorRequirement{{
						Key:      "metadata.name",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"host_2", "host_1"},
					}},
				},
			}},
		},
		{
			name: "nodeLabels and nodeFields: unordered matchExpressions and ordered matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_2_val", "label_1_val"},
						}},
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_1", "host_2"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1",
					Labels: map[string]string{
						"label_1": "label_1_val",
					}}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "label_1",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"label_2_val", "label_1_val"},
					}},
					MatchFields: []v1.NodeSelectorRequirement{{
						Key:      "metadata.name",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"host_1", "host_2"},
					}},
				},
			}},
		},
		{
			name: "nodeLabels and nodeFields: unordered matchExpressions and unordered matchFields",
			args: args{
				nodeSelector: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "label_1",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"label_2_val", "label_1_val"},
						}},
						MatchFields: []v1.NodeSelectorRequirement{{
							Key:      "metadata.name",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"host_2", "host_1"},
						}},
					},
				}},
				node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "host_1",
					Labels: map[string]string{
						"label_1": "label_1_val",
					}}},
			},
			want: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "label_1",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"label_2_val", "label_1_val"},
					}},
					MatchFields: []v1.NodeSelectorRequirement{{
						Key:      "metadata.name",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"host_2", "host_1"},
					}},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _ = MatchNodeSelectorTerms(tt.args.node, tt.args.nodeSelector)
			if !apiequality.Semantic.DeepEqual(tt.args.nodeSelector, tt.want) {
				// fail when tt.args.nodeSelector is deeply modified
				t.Errorf("MatchNodeSelectorTerms() got = %v, want %v", tt.args.nodeSelector, tt.want)
			}
		})
	}
}
