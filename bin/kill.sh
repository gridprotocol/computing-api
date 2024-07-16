#!/bin/bash

ps -ef | grep computing-api | grep -v 'color' | awk '{print $2}' | xargs kill -9 
