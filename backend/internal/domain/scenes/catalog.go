package scenes

const (
	ScenePortrait   = "portrait"
	SceneFestival   = "festival"
	SceneInvitation = "invitation"
	SceneTshirt     = "tshirt"
	ScenePoster     = "poster"
)

var supportedSceneOrder = []string{
	ScenePortrait,
	SceneFestival,
	SceneInvitation,
	SceneTshirt,
	ScenePoster,
}

func SupportedSceneOrder() []string {
	order := make([]string, len(supportedSceneOrder))
	copy(order, supportedSceneOrder)
	return order
}

func IsSupportedScene(sceneKey string) bool {
	for _, item := range supportedSceneOrder {
		if item == sceneKey {
			return true
		}
	}
	return false
}
