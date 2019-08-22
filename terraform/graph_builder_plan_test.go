package terraform

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
)

func TestPlanGraphBuilder_impl(t *testing.T) {
	var _ GraphBuilder = new(PlanGraphBuilder)
}

func TestPlanGraphBuilder_targetModule(t *testing.T) {
	b := &PlanGraphBuilder{
		Config:     testModule(t, "graph-builder-plan-target-module-provider"),
		Components: simpleMockComponentFactory(),
		Schemas:    simpleTestSchemas(),
		Targets: []addrs.Targetable{
			addrs.RootModuleInstance.Child("child2", addrs.NoKey),
		},
	}

	g, err := b.Build(addrs.RootModuleInstance)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	t.Logf("Graph: %s", g.String())

	testGraphNotContains(t, g, "module.child1.provider.test")
	testGraphNotContains(t, g, "module.child1.test_object.foo")
}

const testPlanGraphBuilderStr = `
aws_instance.web
  aws_security_group.firewall
  provider.aws
  var.foo
aws_load_balancer.weblb
  aws_instance.web
  provider.aws
aws_security_group.firewall
  provider.aws
local.instance_id
  aws_instance.web
meta.count-boundary (EachMode fixup)
  aws_instance.web
  aws_load_balancer.weblb
  aws_security_group.firewall
  local.instance_id
  openstack_floating_ip.random
  output.instance_id
  provider.aws
  provider.openstack
  var.foo
openstack_floating_ip.random
  provider.openstack
output.instance_id
  local.instance_id
provider.aws
  openstack_floating_ip.random
provider.aws (close)
  aws_instance.web
  aws_load_balancer.weblb
  aws_security_group.firewall
  provider.aws
provider.openstack
provider.openstack (close)
  openstack_floating_ip.random
  provider.openstack
root
  meta.count-boundary (EachMode fixup)
  provider.aws (close)
  provider.openstack (close)
var.foo
`
const testPlanGraphBuilderForEachStr = `
aws_instance.bar
  provider.aws
aws_instance.bat
  aws_instance.boo
  provider.aws
aws_instance.baz
  provider.aws
aws_instance.boo
  provider.aws
aws_instance.foo
  provider.aws
meta.count-boundary (EachMode fixup)
  aws_instance.bar
  aws_instance.bat
  aws_instance.baz
  aws_instance.boo
  aws_instance.foo
  provider.aws
provider.aws
provider.aws (close)
  aws_instance.bar
  aws_instance.bat
  aws_instance.baz
  aws_instance.boo
  aws_instance.foo
  provider.aws
root
  meta.count-boundary (EachMode fixup)
  provider.aws (close)
`
