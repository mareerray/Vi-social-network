export function resizeAndCropImage(file, size = 512, quality = 0.85) {
  return new Promise((resolve, reject) => {
    try {
      const url = URL.createObjectURL(file)
      const img = new Image()
      img.onload = () => {
        try {
          const iw = img.naturalWidth
          const ih = img.naturalHeight
          // determine square crop
          let sx = 0, sy = 0, s = 0
          if (iw > ih) {
            s = ih
            sx = Math.floor((iw - ih) / 2)
            sy = 0
          } else {
            s = iw
            sx = 0
            sy = Math.floor((ih - iw) / 2)
          }
          const canvas = document.createElement('canvas')
          canvas.width = size
          canvas.height = size
          const ctx = canvas.getContext('2d')
          ctx.drawImage(img, sx, sy, s, s, 0, 0, size, size)
          canvas.toBlob((blob) => {
            URL.revokeObjectURL(url)
            if (!blob) return reject(new Error('Failed to create blob'))
            resolve(blob)
          }, 'image/jpeg', quality)
        } catch (err) {
          URL.revokeObjectURL(url)
          reject(err)
        }
      }
      img.onerror = (e) => {
        URL.revokeObjectURL(url)
        reject(new Error('Image load error'))
      }
      img.src = url
    } catch (err) {
      reject(err)
    }
  })
}
