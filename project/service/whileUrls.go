package service

var (
	W WhileUrls
)

type WhileUrls struct{}

func (WhileUrls) WhileList(url string) bool {
	flag := false
	wls := []string{
		"/login",
		"/logout",
		"/galogin",
	}

	for i := 0; i < len(wls); i++ {
		if wls[i] == url {
			flag = true
		}
	}

	if !flag {
		return flag
	}

	return true
}

func (WhileUrls) OperateWhileList(url string) bool {
	var flag bool
	wls := []string{
		"/perms/list",
		"/role/list",
		"/user/list",
		"/log/list",
		"/role/rolesname",
		"/role/userperms",
		"/user/getinfobyname",
		"/assets/list",
		"/assets/process/update/list",
		"/log/list",
	}

	for i := 0; i < len(wls); i++ {
		if wls[i] == url {
			flag = true
		}
	}

	if !flag {
		return false
	}

	return true
}
