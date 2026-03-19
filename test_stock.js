async function test() {
  try {
    let loginRes = await fetch('http://localhost:8080/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'super_admin', password: 'password' })
    });
	
	if (!loginRes.ok) {
        loginRes = await fetch('http://localhost:8080/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: 'admin', password: 'smsystem_secret_password!123' })
        });
	}

    if (!loginRes.ok) {
        console.log("Failed authentication. Can't test API.");
        return;
    }
    const loginData = await loginRes.json();
    const token = loginData.token;

    console.log("Logged in successfully. Getting products...");
    const prodsRes = await fetch('http://localhost:8080/api/products?all=1', {
      headers: { Authorization: `Bearer ${token}` }
    });
    const prods = await prodsRes.json();
    const p = prods.products.find(x => x.stock === 0) || prods.products[0];
    
    if (!p) return console.log("No products");
    console.log(`Initial product ${p.id} stock: ${p.stock}`);

    console.log("Updating stock to 42...");
    p.stock = 42;
    const updateRes = await fetch(`http://localhost:8080/api/products/${p.id}`, {
      method: 'PUT',
      headers: { 
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}` 
      },
      body: JSON.stringify(p)
    });
    
    const updateData = await updateRes.json();
    console.log("Update response:", updateData.message, "- Stock returned:", updateData.product?.stock);

    console.log("Checking DB directly via API again...");
    const checkRes = await fetch(`http://localhost:8080/api/products/${p.id}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    const checkData = await checkRes.json();
    console.log("Verified stock via API GET:", checkData.product?.stock);

  } catch (err) {
    console.error(err);
  }
}
test();
