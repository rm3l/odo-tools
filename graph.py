#!/usr/bin/env python3

# importing library
import requests
import json
from datetime import date 
import re
import matplotlib.pyplot as plt
import numpy as np
# total no of failure
TotalFailure=0
nooffailure=[]
days=[]
# api-endpoint
URL = "https://search.ci.openshift.org/search"
  
DateAndFailure = {}   
previousFailure=0

# for x in range(336, 0, -24):
for x in range( 24, 336, 24):
    TotalFailure=0
    # defining a params dict for the parameters to be sent to the API
    hour=str(x)+"h"
    PARAMS = {'search':"\\[Fail\\]",'maxAge':hour,'context':"0",'type':"build-log",'name':"pull-ci-redhat-developer-odo-main-",'maxMatches':"5",'maxBytes':"20971520"}
    
    # sending get request and saving the response as response object
    r = requests.get(url = URL, params = PARAMS)

    # extracting data in json format 
    try:
        data = r.json()
    except json.decoder.JSONDecodeError:
        print("Skipping Day ",hour/24)
        continue
    # print('failed CI run',len(data))
    for i in data:
        for j in data[i]:
            for k in data[i][j]:
                #print(k,"\n------------")
                alldata=k["context"]
                TotalFailure=TotalFailure+len(alldata)
                for l in alldata:
                    #print(l)
                    ansi_escape = re.compile(r'''
        \x1B  # ESC
        (?:   # 7-bit C1 Fe (except CSI)
            [@-Z\\-_]
        |     # or [ for CSI, followed by a control sequence
            \[
            [0-?]*  # Parameter bytes
            [ -/]*  # Intermediate bytes
            [@-~]   # Final byte
        )
    ''', re.VERBOSE)
                    result = ansi_escape.sub('',l )
                    #print(result)
    failure=TotalFailure                    
    temp=TotalFailure
    if TotalFailure != 0:
        failure=TotalFailure-previousFailure
    DateAndFailure[int(x/24)]=failure
    days.append(str(int(x/24)))
    nooffailure.append(failure)
    # print('Hours ',x,'\nDays ',int(x/24),'\nfailure ',failure,'\n')
    # print(date.today().strftime("%d/%m/%Y"))
    if temp != 0:
        previousFailure=temp

# print("\n--------------------\n",DateAndFailure)

fig, ax = plt.subplots(figsize=(10, 6))
ax.plot(days,nooffailure, label='linear')
ax.invert_xaxis()
title='Created on '+date.today().strftime("%d/%m/%Y")
plt.title(title)
plt.xlabel('Days')
plt.ylabel('Number of failures')
plt.savefig('temp/graph.png')

print('[graph](temp/graph.png)')