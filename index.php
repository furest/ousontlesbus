<?php
chdir("go-tec");
$filedate = filemtime("map.html");
$now = strtotime("-20 seconds");
if (empty($filedate) || $filedate < strtotime("-20 seconds")){
	shell_exec("./main");
}
include("map.html");
