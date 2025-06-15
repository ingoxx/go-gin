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
		//"/assets/terminal/token",
		//"/assets/terminal/ws",
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
		"/logger/list",
		"/role/rolesname",
		"/role/userperms",
		"/user/getinfobyname",
		"/assets/list",
		"/assets/process/update/list",
		"/logger/list",
		"/assets/program/list",
		"/assets/program/update/list",
		"/log/list",
		"/log/get-login-num",
		"/log/get-run-linux-cmd-num",
		"/assets/terminal/token",
		"/assets/terminal/ws",
		"/assets/run-linux-cmd",
		"/assets/view-system-log",
		"/cluster/list",
		"/assets/ws",
		"/assets/file/ws",
		"/assets/res-vis",
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
