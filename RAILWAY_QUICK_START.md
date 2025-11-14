# Railway Quick Start Guide

## Hızlı Başlangıç

### 1. Railway'e Giriş
- https://railway.app adresine gidin
- GitHub ile giriş yapın

### 2. Backend Deploy

1. **New Project** → **Deploy from GitHub repo**
2. Repository'nizi seçin
3. **Settings** → **Root Directory** → `backend` yazın ve kaydedin
4. **Variables** sekmesinde şu environment variables'ları ekleyin:
   ```
   PORT=8080
   DATA_DIR=/tmp
   ALLOWED_ORIGINS=https://your-frontend-url.railway.app
   ```
   (ALLOWED_ORIGINS'i frontend deploy ettikten sonra güncelleyin)
5. Deploy başlayacak

### 3. Frontend Deploy

1. Aynı project içinde **New Service** → **Deploy from GitHub repo**
2. Aynı repository'yi seçin
3. **Settings** → **Root Directory** → `frontend` yazın ve kaydedin
4. **Variables** sekmesinde şu environment variable'ı ekleyin:
   ```
   VITE_API_URL=https://your-backend-url.railway.app
   ```
   (Backend URL'ini backend service'inizin URL'i ile değiştirin)
5. Deploy başlayacak

### 4. URL'leri Güncelleme

1. Frontend deploy olduktan sonra frontend URL'ini kopyalayın
2. Backend service → Variables → `ALLOWED_ORIGINS` → Frontend URL'i ile güncelleyin
3. Backend'i yeniden deploy edin

### ÖNEMLİ NOTLAR

- ✅ Her service için **Root Directory** mutlaka ayarlanmalı
- ✅ Backend için: `backend`
- ✅ Frontend için: `frontend`
- ✅ Environment variables'ları doğru ayarlayın
- ✅ URL'leri deploy sonrası güncelleyin

### Sorun Giderme

**"Nixpacks build failed" hatası alıyorsanız:**
- Root Directory ayarını kontrol edin
- Service → Settings → Root Directory → Doğru klasör adını yazın
- Service'i yeniden deploy edin

