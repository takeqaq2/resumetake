import zipfile
import os

root = r'f:\project\resumetake'
zip_path = os.path.join(root, 'deploy-r44.zip')

files = [
    r'backend\models\models.go',
    r'backend\services\helpers.go',
    r'frontend\src\routes\admin\+page.svelte',
    r'frontend\src\routes\[lang]\templates\+page.svelte',
    r'frontend\src\routes\+error.svelte',
    r'frontend\src\app.html',
    r'frontend\src\hooks.server.js',
    r'nginx\nginx.conf',
]

with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as z:
    for f in files:
        full = os.path.join(root, f)
        z.write(full, f)

print(f'Created {zip_path} with {len(files)} files')
