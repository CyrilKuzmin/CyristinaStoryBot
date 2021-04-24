import requests
import lxml.html
import random
import json
import sys

URL = 'https://nukadeti.ru/skazki/gusi_lebedi'

def translit(text):
    slovar = {'а':'a','б':'b','в':'v','г':'g','д':'d','е':'e','ё':'yo',
      'ж':'zh','з':'z','и':'i','й':'i','к':'k','л':'l','м':'m','н':'n',
      'о':'o','п':'p','р':'r','с':'s','т':'t','у':'u','ф':'f','х':'h',
      'ц':'c','ч':'ch','ш':'sh','щ':'sch','ъ':'','ы':'y','ь':'','э':'e',
      'ю':'u','я':'ya', 'А':'A','Б':'B','В':'V','Г':'G','Д':'D','Е':'E','Ё':'YO',
      'Ж':'ZH','З':'Z','И':'I','Й':'I','К':'K','Л':'L','М':'M','Н':'N',
      'О':'O','П':'P','Р':'R','С':'S','Т':'T','У':'U','Ф':'F','Х':'H',
      'Ц':'C','Ч':'CH','Ш':'SH','Щ':'SCH','Ъ':'','Ы':'y','Ь':'','Э':'E',
      'Ю':'U','Я':'YA',',':'','?':'',' ':'_','~':'','!':'','@':'','#':'',
      '$':'','%':'','^':'','&':'','*':'','(':'',')':'','-':'','=':'','+':'',
      ':':'',';':'','<':'','>':'','\'':'','"':'','\\':'','/':'','№':'',
      '[':'',']':'','{':'','}':'','ґ':'','ї':'', 'є':'','Ґ':'g','Ї':'i',
      'Є':'e', '—':''}
    for key in slovar:
        text = text.replace(key, slovar[key])
    return text


response = requests.get(URL)
tree = lxml.html.fromstring(response.text)
TITLE = tree.xpath('/html/body/div[1]/div[3]/div[1]/div[1]/div[1]/h1')[0].text
TITLE_IMAGE = "https://nukadeti.ru" + tree.xpath('/html/body/div[1]/div[3]/div[1]/div[1]/div[3]/img')[0].items()[0][1]
raw_captions = []
images_urls = []

content_path = '/html/body/div[1]/div[3]/div[1]/div[1]/div[4]/div[2]'
p_cnt = 1
while True:
    p_path = content_path + f'/p[{p_cnt}]/'
    try:
        # Пытаемся дернуть текст 
        br_cnt = 1
        while True:
            br_path = p_path + f'text()[{br_cnt}]'
            data = tree.xpath(br_path)
            br_cnt += 1
            if not data:
                raw_captions.append('PICTURE_IS_HERE')
                break
            raw_captions.append(data)
    except Exception as e:
        pass
    
    try:
        # Пытаемся дернуть картинки
        data = tree.xpath(f'{p_path}img')[0]
        images_urls.append("https://nukadeti.ru" + data.items()[1][1])

    except Exception as e:
        pass
    p_cnt += 1
    if p_cnt == 1000:
        break

captions = []
for i in range(len(raw_captions)):
    if raw_captions[i] != 'PICTURE_IS_HERE':
        captions.append(raw_captions[i][0].replace('\r', ''))
        continue
    if raw_captions[i] == 'PICTURE_IS_HERE' and raw_captions[i-1] == 'PICTURE_IS_HERE':
        continue
    if raw_captions[i] == 'PICTURE_IS_HERE' and raw_captions[i-1] != 'PICTURE_IS_HERE':
        captions.append('PICTURE_IS_HERE')

FINAL_CAPTIONS = [value for value in ''.join(captions).split('PICTURE_IS_HERE') if value]
#print(len(FINAL_CAPTIONS))
if len(FINAL_CAPTIONS) > len(images_urls):
    elems_to_union = len(FINAL_CAPTIONS) - len(images_urls)
    positions = []
    for i in range(elems_to_union):
        minimal = 0
        for c in range(len(FINAL_CAPTIONS)):
            if len(FINAL_CAPTIONS[c]) < len(FINAL_CAPTIONS[minimal]):
                minimal = c 
        if minimal != 0:
            FINAL_CAPTIONS[minimal-1] += FINAL_CAPTIONS[minimal]
            FINAL_CAPTIONS.pop(minimal)
        else:
            FINAL_CAPTIONS[minimal] += FINAL_CAPTIONS[minimal+1]
            FINAL_CAPTIONS.pop(minimal+1)

title_image_data = requests.get(TITLE_IMAGE).content
format = TITLE_IMAGE.split('.')[-1]
TITLE_IMAGE_FILENAME = f'{translit(TITLE)}_0.{format}'
with open(TITLE_IMAGE_FILENAME, 'wb') as handler:
    handler.write(title_image_data)

IMAGES = []

for i in range(len(images_urls)):
    url = images_urls[i]
    format = url.split('.')[-1]
    img_data = requests.get(url).content
    filename = f'{translit(TITLE)}_{i+1}.{format}'
    with open(filename, 'wb') as handler:
        handler.write(img_data)
    IMAGES.append(filename)

result = {
    'ID': random.randint(1,9999),
    'Title': TITLE,
    'Content': [
        {
            'Image': TITLE_IMAGE_FILENAME,
            'Caption': f'*{TITLE}*'
        }
    ]
}

for i in range(len(FINAL_CAPTIONS)):
    result['Content'].append(
        {
            'Image': IMAGES[i],
            'Caption': FINAL_CAPTIONS[i]
        }
    )

with open(f'{translit(TITLE)}.json', 'w',  encoding='utf8') as file:
    json.dump(result, file, indent=4, ensure_ascii=False)