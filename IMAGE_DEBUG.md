## 🔍 Debug Image Payload

Untuk enable image processing, kita perlu:

1. **Check GOWA image payload format**
   - Tambah logging untuk image messages
   - Lihat struktur JSON yang dikirim GOWA

2. **Update IncomingMessage struct**
   - Tambah field untuk media (image_url, media_type, etc.)

3. **Implement image download & OCR**
   - Download image dari GOWA
   - Process dengan Vision AI
   - Extract transaction data

## 📋 Next Steps:

1. Add debug logging untuk image messages
2. Deploy dan kirim foto struk lagi
3. Check logs untuk payload structure
4. Update struct dan implement image handling

Kirim foto struk sekali lagi, nanti kita lihat payload nya di logs!
