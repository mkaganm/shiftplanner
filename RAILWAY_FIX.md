# Railway Deployment Hatası Çözümü

## Sorun
Railway "Railpack could not determine how to build the app" hatası veriyor.

## Çözüm Adımları

### 1. Railway Dashboard'da Root Directory Ayarlama

**Backend Service için:**
1. Railway dashboard'a gidin
2. Backend service'inize tıklayın
3. **Settings** sekmesine gidin
4. **Root Directory** alanını bulun
5. `backend` yazın (sadece `backend`, başka bir şey yazmayın)
6. **Save** butonuna tıklayın
7. Service'i yeniden deploy edin

**Frontend Service için:**
1. Frontend service'inize tıklayın
2. **Settings** sekmesine gidin
3. **Root Directory** alanına `frontend` yazın
4. **Save** butonuna tıklayın
5. Service'i yeniden deploy edin

### 2. Alternatif Çözüm: Service'i Silip Yeniden Oluşturma

Eğer yukarıdaki adımlar işe yaramazsa:

1. Mevcut service'i silin
2. Yeni service oluştururken:
   - **Deploy from GitHub repo** seçin
   - Repository'nizi seçin
   - **Root Directory** alanına hemen `backend` veya `frontend` yazın
   - Deploy edin

### 3. Railway.json Dosyalarını Kontrol Edin

`backend/railway.json` ve `frontend/railway.json` dosyaları doğru yerde olmalı.

### 4. Environment Variables

Deploy olduktan sonra environment variables'ları ekleyin:

**Backend:**
```
PORT=8080
DATA_DIR=/tmp
ALLOWED_ORIGINS=https://your-frontend-url.railway.app
```

**Frontend:**
```
VITE_API_URL=https://your-backend-url.railway.app
```

## Önemli Notlar

- ✅ Root Directory ayarı **mutlaka** yapılmalı
- ✅ Her service için ayrı ayrı ayarlayın
- ✅ Root Directory sadece klasör adı olmalı (`backend` veya `frontend`)
- ✅ `/backend` veya `./backend` gibi path'ler kullanmayın
- ✅ Ayarları kaydettikten sonra service'i yeniden deploy edin

## Hala Sorun Varsa

1. Railway logs'u kontrol edin
2. Root Directory ayarının doğru olduğundan emin olun
3. Service'i silip yeniden oluşturmayı deneyin

