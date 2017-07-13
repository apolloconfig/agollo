package component

type AbsComponent interface {
	Start()
}


func StartRefreshConfig(component AbsComponent)  {
	component.Start()
}