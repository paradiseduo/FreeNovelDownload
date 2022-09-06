import json
import random
import math
import getopt
import requests
import sys
import signal
import re
from requests.utils import requote_uri


def printUse():
    print('''
    Usage:      
        python3 downloader.py -n 斗破苍穹 -p proxyList

        -h show help
        -n --name <Book Name>
        -p --proxy <Proxy File Path>
    ''')


def handle_SIGINT(signum, frame):
    exit(-1)


def remove_upprintable_chars(s):
    """移除所有不可见字符"""
    return ''.join(x for x in s if x.isprintable())


def remove(str, num):
    return str[:num] + str[num + 1:]


def json_loads(data_str):
    try:
        result = json.loads(data_str)
        return result
    except Exception as e:
        error_index = re.findall(r"char (\d+)\)", str(e))
        if error_index:
            newStr2 = remove(data_str, int(error_index[0]))
            return json_loads(newStr2)


def sendRequest(client, url, proxys):
    if len(proxys) > 0:
        proxy = str(random.choice(proxys)).lower()
        if proxy.startswith('https://'):
            client.proxies.update({'https': proxy})
        if proxy.startswith('http://'):
            client.proxies.update({'http': proxy})
        if proxy.startswith('socks5://'):
            client.proxies.update({'socks5': proxy})
    try:
        r = client.get(url, timeout=5)
        r.encoding = 'utf-8'
        return r.text
    except requests.exceptions.RequestException as e:
        if len(proxys) > 0:
            proxys.remove(proxy)
        return sendRequest(client, url, proxys)


def main(argv):
    signal.signal(signal.SIGINT, handle_SIGINT)
    name = ''
    proxyFile = ''
    try:
        opts, args = getopt.getopt(argv, "hn:p:", ["name=", "proxy="])
    except getopt.GetoptError:
        printUse()
        sys.exit(2)

    for (opt, arg) in opts:
        if opt == "-h":
            printUse()
            sys.exit()
        elif opt in ("-n", "--name"):
            name = arg
        elif opt in ("-p", "--proxy"):
            proxyFile = arg

    if name == '':
        printUse()
        exit(0)

    proxys = []
    if proxyFile != '':
        with open(proxyFile, 'r') as f:
            for line in f:
                proxys.append(line.strip())
    headers = {
        "Accept": "*/*",
        "Accept-Language": "zh-Hans-CN;q=1",
        "Connection": "keep-alive",
        "Accept-Encoding": "gzip, deflate, br",
        "User-Agent": "",
    }

    client = requests.Session()
    client.keep_alive = True
    client.headers = headers
    client.encoding = 'utf-8'

    saveDic = {}
    searchURL = f'https://souxs.leeyegy.com/search.aspx?key={name}&siteid=app2'
    index = 0
    for i in range(0, 10):
        searchURL = requote_uri(f'{searchURL}&page={i}')
        res = sendRequest(client, searchURL, proxys)
        jsonData = json_loads(res)
        datas = jsonData['data']
        if len(datas) == 0:
            break
        for item in datas:
            saveDic[str(index)] = {'name': item['Name'], 'id': item['Id'],
                                   'sort': math.ceil(float(item['Id']) / 1000.0)}
            print(index, item['Name'], item['Author'])
            index += 1

    iii = input('请输入想要下载的序号：')
    info = saveDic[iii]
    sort = info['sort']
    id = info['id']
    baseurl = f'https://downbakxs.apptuxing.com/BookFiles/Html/{str(sort)}/{id}'
    res = sendRequest(client, baseurl, proxys)
    jsonData = json_loads(remove_upprintable_chars(res.replace(',]', ']').replace(',}', '}')))
    bookName = jsonData['data']['name']
    bookList = jsonData['data']['list']

    with open(bookName, 'a') as f:
        for item in bookList:
            ccName = item['name']
            f.write('\n\n' + remove_upprintable_chars(ccName) + '\n')
            ccList = item['list']
            for page in ccList:
                res1 = sendRequest(client, baseurl + '/' + str(page['id']) + '.html', proxys)
                if res1 == "":
                    res1 = sendRequest(client, baseurl + '/' + str(page['id']) + '.html', proxys)
                jsonData2 = json_loads(res1[1:])
                capName = page['name']
                content = jsonData2['data']['content'].replace('\r\n', '∑').replace('\n', '')
                f.write('\n\n' + remove_upprintable_chars(capName) + '\n')
                f.write(remove_upprintable_chars(content).replace('∑', '\r\n　　'))
                print(capName)


if __name__ == '__main__':
    main(sys.argv[1:])
