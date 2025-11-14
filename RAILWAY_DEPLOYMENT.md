# Railway Deployment Guide

Bu projeyi Railway'de deploy etmek için adım adım rehber.

## Gereksinimler

- Railway hesabı (https://railway.app - GitHub ile giriş yapabilirsiniz)
- GitHub repository'de kodunuzun olması

## Deployment Adımları

### 1. Railway'e Giriş Yapın

1. https://railway.app adresine gidin
2. "Login" butonuna tıklayın
3. GitHub hesabınızla giriş yapın

### 2. Backend Service Oluşturun

1. Railway dashboard'da "New Project" butonuna tıklayın
2. "Deploy from GitHub repo" seçeneğini seçin
3. Repository'nizi seçin
4. "Root Directory" olarak `backend` klasörünü seçin
5. Railway otomatik olarak Go projesini algılayacak

#### Backend Environment Variables

Railway dashboard'da backend service'inizde şu environment variables'ları ekleyin:

```
PORT=8080
DATA_DIR=/tmp
ALLOWED_ORIGINS=https://your-frontend-url.railway.app
```

**ÖNEMLİ:** `ALLOWED_ORIGINS` değerini frontend URL'iniz ile değiştirin (frontend deploy ettikten sonra).

### 3. Frontend Service Oluşturun

1. Railway dashboard'da "New Service" butonuna tıklayın
2. "Deploy from GitHub repo" seçeneğini seçin
3. Aynı repository'yi seçin
4. "Root Directory" olarak `frontend` klasörünü seçin
5. Railway otomatik olarak Node.js projesini algılayacak

#### Frontend Environment Variables

Railway dashboard'da frontend service'inizde şu environment variable'ı ekleyin:

```
VITE_API_URL=https://your-backend-url.railway.app
```

**ÖNEMLİ:** `VITE_API_URL` değerini backend URL'iniz ile değiştirin.

### 4. Build Ayarları

#### Backend Build Settings

Railway otomatik olarak Go projesini algılayacak. Eğer sorun olursa:

- **Build Command:** `go build -o server ./cmd/server`
- **Start Command:** `./server`

#### Frontend Build Settings

Railway otomatik olarak Node.js projesini algılayacak. Eğer sorun olursa:

- **Build Command:** `npm install && npm run build`
- **Start Command:** `npm run preview`

### 5. Port Ayarları

Railway otomatik olarak port'u ayarlar. Her iki service için de `PORT` environment variable'ı otomatik olarak set edilir.

### 6. Database

Backend SQLite kullanıyor. Railway'de `/tmp` dizini kullanılabilir. `DATA_DIR=/tmp` environment variable'ı ile database `/tmp` dizinine kaydedilir.

**NOT:** Railway'de `/tmp` dizini geçicidir. Kalıcı veri için Railway'in PostgreSQL servisini kullanabilirsiniz (gelecekte).

### 7. CORS Ayarları

Backend'de `ALLOWED_ORIGINS` environment variable'ını frontend URL'iniz ile güncelleyin:

```
ALLOWED_ORIGINS=https://your-frontend-url.railway.app
```

### 8. Deploy

1. Her iki service için de Railway otomatik olarak deploy edecek
2. Deploy tamamlandıktan sonra her service için bir URL alacaksınız
3. Frontend URL'ini backend'deki `ALLOWED_ORIGINS` ile güncelleyin
4. Backend URL'ini frontend'deki `VITE_API_URL` ile güncelleyin
5. Her iki service'i de yeniden deploy edin

## Troubleshooting

### Backend başlamıyor

- `PORT` environment variable'ının set olduğundan emin olun
- Railway logs'u kontrol edin
- `DATA_DIR` environment variable'ını `/tmp` olarak ayarlayın

### Frontend API'ye bağlanamıyor

- `VITE_API_URL` environment variable'ının doğru olduğundan emin olun
- Backend URL'inin doğru olduğundan emin olun
- CORS ayarlarını kontrol edin

### CORS hatası

- Backend'deki `ALLOWED_ORIGINS` environment variable'ını frontend URL'iniz ile güncelleyin
- URL'lerin `https://` ile başladığından emin olun

## Ücretsiz Tier Limitleri

Railway'in ücretsiz tier'ı:
- $5 kredi/ay
- Küçük projeler için yeterli
- Uyku modu: 30 dakika kullanılmazsa uykuya geçer (ilk istekte uyanır)

## İpuçları

1. Her iki service'i de aynı Railway project'inde tutun
2. Environment variables'ları doğru ayarlayın
3. Deploy sonrası URL'leri güncelleyin
4. Railway logs'u takip edin

## Destek

Sorun yaşarsanız Railway'in dokümantasyonunu kontrol edin: https://docs.railway.app

