<?php
chdir("go-tec");
$filedate = filemtime("map.html");
$now = strtotime("-20 seconds");
if ($filedate < strtotime("-20 seconds")){
	shell_exec("./main");
}
include("map.html");
