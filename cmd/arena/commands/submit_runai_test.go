package commands

import (
	"testing"
)

var (
	logsString = "Executing the command: jupyter notebook\n" +
		"Writing notebook server cookie secret to /home/jovyan/.local/share/jupyter/runtime/notebook_cookie_secret\n" +
		"JupyterLab extension loaded from /opt/conda/lib/python3.7/site-packages/jupyterlab\n" +
		" JupyterLab application directory is /opt/conda/share/jupyter/lab\n" +
		" Serving notebooks from local directory: /home/jovyan\n" +
		" The Jupyter Notebook is running at:\n" +
		" http://or-test5-0:8888/?token=61965193b93fca57bf64348b7e2be160af182a41134662e3\n" +
		"  or http://127.0.0.1:8888/?token=61965193b93fca57bf64348b7e2be160af182a41134662e3\n" +
		" Use Control-C to stop this server and shut down all kernels (twice to skip confirmation).\n" +
		"\n" +
		"\n" +
		"	To access the notebook, open this file in a browser:\n" +
		"		file:///home/jovyan/.local/share/jupyter/runtime/nbserver-6-open.html\n" +
		"	Or copy and paste one of these URLs:\n" +
		"		http://or-test5-0:8888/?token=61965193b93fca57bf64348b7e2be160af182a41134662e3\n" +
		"	 or http://127.0.0.1:8888/?token=61965193b93fca57bf64348b7e2be160af182a41134662e3 \n"
)

func TestJupyterTokenRegex(t *testing.T) {
	expected := "61965193b93fca57bf64348b7e2be160af182a41134662e3"
	res, err := getTokenFromJupyterLogs(logsString)
	if err != nil || res != expected {
		t.Errorf("Strings dont match expected: %s, result: %s", expected, res)
	}
}

func TestStringNotFound(t *testing.T) {
	res, err := getTokenFromJupyterLogs("some other string")
	if res != "" || err == nil {
		t.Errorf("Did not return error when token not found")
	}
}
