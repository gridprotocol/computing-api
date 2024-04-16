#!/bin/bash

ps -ef | grep computing-gw | grep -v 'color' | awk '{print $2}' | xargs kill -9 
