ps -ef | grep gateway | grep -v 'color' | awk '{print $2}' | xargs kill -9 
nohup ./gateway > log 2>&1 &