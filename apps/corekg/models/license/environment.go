package license

import (
	"context"
	"fmt"
	"os"

	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Environment 定义了与特定环境交互的接口
type Environment interface {
	GetUID(ctx context.Context) (string, error)
	GetRawLicense(ctx context.Context) (string, error)
}
type EnvType string

// --- Kubernetes Environment 实现 ---

var (
	EnvTypeKubernetes EnvType = "kubernetes"
	EnvTypePhysical   EnvType = "physical"
)

// KubernetesEnvironment 实现了在 K8s 环境中获取数据
type KubernetesEnvironment struct {
	LicensePath string
	//PublicKeyPath string
}

func (k *KubernetesEnvironment) GetUID(ctx context.Context) (string, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to create in-cluster config: %w", err)
	}
	client, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	namespace, err := client.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get kube-system namespace: %w", err)
	}
	return string(namespace.UID), nil
}

func (k *KubernetesEnvironment) GetRawLicense(ctx context.Context) (string, error) {
	logs.DebugContextf(ctx, "Get license from core_setting corekg:raw_license")
	lic, err := settings.GetValue("corekg", "raw_license")
	if err != nil {
		logs.WarnContextf(ctx, "Get coresettings'RawLicense error: %v", err)
		logs.DebugContextf(ctx, "GetRawLicense from %v", k.LicensePath)
		bts, err := os.ReadFile(k.LicensePath)
		if err != nil {
			logs.ErrorContextf(ctx, "Failed to read license file %s: %v", k.LicensePath, err)
			return "", fmt.Errorf("failed to read license file %s: %w", k.LicensePath, err)
		}
		return string(bts), nil
	}
	return lic, nil

}

// PhysicalEnvironment 实现了在物理机（非 K8s）环境中获取数据
type PhysicalEnvironment struct {
	LicensePath string
}

func (p *PhysicalEnvironment) GetUID(ctx context.Context) (string, error) {
	// TODO: 生成物理机的稳定唯一标识，暂以占位值返回，避免启动报错。
	return "physical-uid", nil
}

// GetRawLicense 优先从 core_settings 读取，未配置时回退到本地 license 文件
func (p *PhysicalEnvironment) GetRawLicense(ctx context.Context) (string, error) {
	logs.DebugContextf(ctx, "Get license from core_setting corekg:raw_license")
	lic, err := settings.GetValue("corekg", "raw_license")
	if err == nil {
		return lic, nil
	}
	logs.WarnContextf(ctx, "Get coresettings'RawLicense error: %v", err)
	logs.DebugContextf(ctx, "GetRawLicense from %v", p.LicensePath)
	bts, err := os.ReadFile(p.LicensePath)
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to read license file %s: %v", p.LicensePath, err)
		return "", fmt.Errorf("failed to read license file %s: %w", p.LicensePath, err)
	}
	return string(bts), nil
}

// NewEnvironment 根据环境类型创建对应的 Environment 实例
func NewEnvironment(envType EnvType) (Environment, error) {
	switch envType {
	case EnvTypeKubernetes:
		return &KubernetesEnvironment{
			LicensePath: "/etc/sys/license/license.dat",
		}, nil
	case EnvTypePhysical:
		return &PhysicalEnvironment{
			LicensePath: "/etc/sys/license/license.dat",
		}, nil
	default:
		// 未明确指定环境类型（如本地调试）时回退到物理机环境，避免未知环境类型报错。
		logs.Warnf("unknown environment type %q, fallback to physical", envType)
		return &PhysicalEnvironment{
			LicensePath: "/etc/sys/license/license.dat",
		}, nil
	}
}
