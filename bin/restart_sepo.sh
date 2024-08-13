ps -ef | grep computing-api | grep -v 'color' | awk '{print $2}' | xargs kill -9 
nohup ./computing-api daemon run --chain sepo --pw 123123> log 2>&1 &