import json
import random
import math
import getopt
import requests
import sys
from requests.utils import requote_uri


def printUse():
    print('''
    Usage:      
        python3 downloader.py -n 斗破苍穹 -p proxyList

        -h show help
        -n --name <Book Name>
        -p --proxy <Proxy File Path>
    ''')


def remove_upprintable_chars(s):
    """移除所有不可见字符"""
    return ''.join(x for x in s if x.isprintable())


def main(argv):
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
    for i in range(0, 100):
        searchURL = requote_uri(f'{searchURL}&page={i}')
        res = client.get(searchURL)
        res.encoding = 'utf-8'
        jsonData = json.loads(res.text)
        datas = jsonData['data']
        if len(datas) == 0:
            break
        for item in datas:
            saveDic[str(index)] = {'name': item['Name'], 'id': item['Id'], 'sort': math.ceil(float(item['Id'])/1000.0)}
            print(index, item['Name'], item['Author'])
            index += 1

    iii = input('请输入想要下载的序号：')
    info = saveDic[iii]
    sort = info['sort']
    id = info['id']
    baseurl = f'https://downbakxs.apptuxing.com/BookFiles/Html/{str(sort)}/{id}'
    res = client.get(baseurl)
    res.encoding = 'utf-8'
    jsonData = json.loads(remove_upprintable_chars(res.text.replace(',]', ']').replace(',}', '}')))
    bookName = jsonData['data']['name']
    bookList = jsonData['data']['list']
    with open(bookName, 'a') as f:
        for item in bookList:
            ccName = item['name']
            f.write('\n\n' + remove_upprintable_chars(ccName) + '\n')
            ccList = item['list']
            for page in ccList:
                if len(proxys) > 0:
                    proxy = str(random.choice(proxys)).lower()
                    if proxy.startswith('https://'):
                        client.proxies.update({'https': proxy})
                    if proxy.startswith('http://'):
                        client.proxies.update({'http': proxy})
                    if proxy.startswith('socks5://'):
                        client.proxies.update({'socks5': proxy})
                res1 = client.get(baseurl + '/' + str(page['id']) + '.html')
                res1.encoding = 'utf-8'
                jsonData2 = json.loads(res1.text[1:])
                capName = page['name']
                content = jsonData2['data']['content'].replace('\r\n　　', '∑').replace('\n', '').replace('　　', '')
                f.write('\n\n' + remove_upprintable_chars(capName) + '\n')
                f.write(remove_upprintable_chars(content).replace('∑', '\r\n　　'))
                print(capName)


if __name__ == '__main__':
        main(sys.argv[1:])
