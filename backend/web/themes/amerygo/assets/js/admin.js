const AdminApp = {
    products: [],
    // Default Config (Overridden by Settings)
    config: {
        cloudName: 'demo',
        uploadPreset: 'docs_upload_example_preset'
    },

    init: async function() {
        this.loadSettings();
        this.loadProducts();
        this.renderTable();
        
        // Form Handler
        document.getElementById('product-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveProduct();
        });

        // Image Upload Handler
        document.getElementById('p-image-file').addEventListener('change', (e) => {
            this.handleImageUpload(e.target.files[0]);
        });
    },

    loadSettings: function() {
        const stored = localStorage.getItem('amerygo_settings');
        if (stored) {
            const settings = JSON.parse(stored);
            // Merge all settings into config
            this.config = { ...this.config, ...settings };
        }
    },

    loadProducts: function() {
        const stored = localStorage.getItem('amerygo_products');
        if (stored) {
            this.products = JSON.parse(stored);
        } else {
            // Initial load from JSON if LS empty (should have been handled by frontend, but safety first)
            // For admin, we start empty or assume frontend synced it. 
            // Better to fetch if empty.
            fetch('../assets/data/products.json')
                .then(res => res.json())
                .then(data => {
                    this.products = data;
                    localStorage.setItem('amerygo_products', JSON.stringify(data));
                    this.renderTable();
                })
                .catch(() => this.products = []);
        }
    },

    renderTable: function() {
        const tbody = document.getElementById('product-list');
        tbody.innerHTML = '';
        
        this.products.forEach(p => {
            const tr = document.createElement('tr');
            const imgHtml = p.image 
                ? `<img src="${p.image}" style="width: 50px; height: 50px; object-fit: cover; border-radius: 4px;">`
                : '<div style="width:50px; height:50px; background:#eee;"></div>';
                
            tr.innerHTML = `
                <td>${imgHtml}</td>
                <td>${p.name}</td>
                <td>${p.category}</td>
                <td>Rp ${p.price.toLocaleString('id-ID')}</td>
                <td>
                    <button class="action-btn btn-edit" onclick="AdminApp.editProduct(${p.id})"><i class="fas fa-edit"></i></button>
                    <button class="action-btn btn-delete" onclick="AdminApp.deleteProduct(${p.id})"><i class="fas fa-trash"></i></button>
                </td>
            `;
            tbody.appendChild(tr);
        });
    },

    openModal: function() {
        document.getElementById('product-modal').style.display = 'flex';
        document.getElementById('product-form').reset();
        document.getElementById('p-id').value = '';
        document.getElementById('modal-title').innerText = 'Tambah Produk';
        document.getElementById('preview-img').style.display = 'none';
    },

    closeModal: function() {
        document.getElementById('product-modal').style.display = 'none';
    },

    editProduct: function(id) {
        const p = this.products.find(x => x.id === id);
        if (!p) return;

        this.openModal();
        document.getElementById('modal-title').innerText = 'Edit Produk';
        document.getElementById('p-id').value = p.id;
        document.getElementById('p-name').value = p.name;
        document.getElementById('p-category').value = p.category;
        document.getElementById('p-price').value = p.price;
        document.getElementById('p-desc').value = p.description || '';
        document.getElementById('p-image-url').value = p.image || '';
        
        if (p.image) {
            const preview = document.getElementById('preview-img');
            preview.src = p.image;
            preview.style.display = 'block';
        }
    },

    deleteProduct: function(id) {
        if(confirm('Yakin ingin menghapus produk ini?')) {
            this.products = this.products.filter(p => p.id !== id);
            this.saveToStorage();
            this.renderTable();
        }
    },

    saveProduct: function() {
        const id = document.getElementById('p-id').value;
        const name = document.getElementById('p-name').value;
        const category = document.getElementById('p-category').value;
        const price = parseInt(document.getElementById('p-price').value);
        const desc = document.getElementById('p-desc').value;
        const image = document.getElementById('p-image-url').value;

        if (id) {
            // Update
            const index = this.products.findIndex(p => p.id == id);
            if (index !== -1) {
                this.products[index] = { ...this.products[index], name, category, price, description: desc, image };
            }
        } else {
            // Create
            const newId = this.products.length > 0 ? Math.max(...this.products.map(p => p.id)) + 1 : 1;
            this.products.push({ id: newId, name, category, price, description: desc, image, gallery: [], rating: 5 });
        }

        this.saveToStorage();
        this.closeModal();
        this.renderTable();
    },

    saveToStorage: function() {
        localStorage.setItem('amerygo_products', JSON.stringify(this.products));
    },

    handleImageUpload: async function(file) {
        if (!file) return;

        const statusDiv = document.getElementById('upload-status');
        
        // Check Provider
        const provider = this.config.storageProvider || 'cloudinary';
        
        if (provider === 'minio') {
            await this.uploadToMinio(file, statusDiv);
        } else {
            await this.uploadToCloudinary(file, statusDiv);
        }
    },

    uploadToCloudinary: async function(file, statusDiv) {
        statusDiv.innerText = 'Mengupload ke Cloudinary...';
        statusDiv.style.color = 'blue';

        const formData = new FormData();
        formData.append('file', file);
        formData.append('upload_preset', this.config.uploadPreset);
        
        const url = `https://api.cloudinary.com/v1_1/${this.config.cloudName}/image/upload`;

        try {
            const response = await fetch(url, {
                method: 'POST',
                body: formData
            });
            const data = await response.json();
            
            if (data.secure_url) {
                this.updateImageField(data.secure_url);
                statusDiv.innerText = 'Upload sukses!';
                statusDiv.style.color = 'green';
            } else {
                throw new Error('Upload failed');
            }
        } catch (error) {
            console.error('Error uploading:', error);
            statusDiv.innerText = 'Upload gagal. Cek koneksi atau config.';
            statusDiv.style.color = 'red';
        }
    },

    uploadToMinio: async function(file, statusDiv) {
        statusDiv.innerText = 'Mengupload ke Minio...';
        statusDiv.style.color = 'blue';

        const c = this.config;
        if(!c.minioEndpoint || !c.minioAccessKey || !c.minioSecretKey || !c.minioBucket) {
             statusDiv.innerText = 'Config Minio belum lengkap! Cek Pengaturan.';
             statusDiv.style.color = 'red';
             return;
        }

        try {
            const useSSL = c.minioUseSSL !== false;
            const protocol = useSSL ? 'https' : 'http';
            const defaultPort = useSSL ? 443 : 80;
            const port = c.minioPort || defaultPort;
            
            // Build endpoint URL for AWS SDK
            const portStr = (port !== 443 && port !== 80) ? `:${port}` : '';
            const endpointUrl = `${protocol}://${c.minioEndpoint}${portStr}`;
            
            const s3 = new AWS.S3({
                endpoint: endpointUrl,
                accessKeyId: c.minioAccessKey,
                secretAccessKey: c.minioSecretKey,
                s3ForcePathStyle: true,
                signatureVersion: 'v4'
            });

            // Generate unique filename
            const fileName = Date.now() + '-' + file.name.replace(/\s+/g, '-');
            
            const params = {
                Bucket: c.minioBucket,
                Key: fileName,
                Body: file,
                ContentType: file.type
            };

            const self = this;
            s3.upload(params, (err, data) => {
                if (err) {
                    console.error('Minio Upload Error:', err);
                    statusDiv.innerText = `Upload Minio Gagal: ${err.message}`;
                    statusDiv.style.color = 'red';
                } else {
                    // Construct Public URL (no port for standard ports)
                    const fileUrl = `${protocol}://${c.minioEndpoint}${portStr}/${c.minioBucket}/${fileName}`;
                    self.updateImageField(fileUrl);
                    statusDiv.innerText = 'Upload ke Minio sukses!';
                    statusDiv.style.color = 'green';
                }
            });

        } catch (error) {
            console.error('Minio Upload Error:', error);
            statusDiv.innerText = `Upload Minio Gagal: ${error.message}`;
            statusDiv.style.color = 'red';
        }
    },

    updateImageField: function(url) {
        document.getElementById('p-image-url').value = url;
        const preview = document.getElementById('preview-img');
        preview.src = url;
        preview.style.display = 'block';
    },

    downloadJSON: function() {
        const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(this.products, null, 2));
        const downloadAnchorNode = document.createElement('a');
        downloadAnchorNode.setAttribute("href", dataStr);
        downloadAnchorNode.setAttribute("download", "products.json");
        document.body.appendChild(downloadAnchorNode); // required for firefox
        downloadAnchorNode.click();
        downloadAnchorNode.remove();
    }
};
