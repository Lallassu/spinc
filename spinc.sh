#!/bin/bash
screen -dm ./ngrok/ngrok http 2601
sleep 2
pid=`ps a |grep "ngrok http 2601" |grep -v login|grep -v grep |awk '{print $1}'`
echo "Ngrok PID: $pid"
grok=`curl -s -i http://127.0.0.1:4040/inspect/http |grep ngrok.io |sed -n 's#.*\(https://[^\]*\).*#\1#;p'`
echo "Ngrok URL: $grok"
./spinc $grok
kill -3 $pid
