package gatewayV2

type apiCoins struct {
	api    *apiCreator
	intgs  map[string]*integrationCreator
	routes map[string]*routeCreator
	stages map[string]*stageCreator
}
