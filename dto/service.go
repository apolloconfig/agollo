package dto

type ServiceDTO struct {
	appName     string
	instanceId  string
	homepageUrl string
}

func CreateServiceDTO(appName string,
	instanceId string,
	homepageUrl string) (service *ServiceDTO) {
	service = &ServiceDTO{
		appName:     appName,
		instanceId:  instanceId,
		homepageUrl: homepageUrl,
	}
	return
}
