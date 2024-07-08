#!/bin/bash

ps -ef | grep gateway | grep -v 'color' | awk '{print $2}' | xargs kill -9 
