ps -ef | grep computing-api | grep -v 'color' | awk '{print $2}' | xargs kill -9 
nohup ./computing-api -chain local > log 2>&1 &