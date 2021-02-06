package boot

import "code-platform/app/service/component"

func init() {
	//初始化casbin的adapter
	component.InitCasbin()

	//初始化minio
	component.InitMinioUtil()
}
