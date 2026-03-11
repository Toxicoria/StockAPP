<script>
  let mensajeServidor = "Esperando para conectar...";
  let estado = "neutral";

  async function hacerPing() {
    mensajeServidor = "Viajando por el Sidecar...";
    estado = "neutral";
    
    try {
      const respuesta = await fetch("http://localhost:9090/api/ping");
      
      if (respuesta.ok) {
        const datos = await respuesta.json();
        mensajeServidor = datos.mensaje;
        estado = "exito";
      } else {
        mensajeServidor = "El servidor respondió con error: " + respuesta.status;
        estado = "error";
      }
    } catch (error) {
      mensajeServidor = "Fallo de red. ¿Está corriendo el Sidecar?";
      estado = "error";
      console.error(error);
    }
  }
</script>

<main>
  <h1>Control de Stock 📦</h1>
  <p>Panel de Prueba de Conexión</p>

  <div class="panel">
    <button on:click={hacerPing}>
      Lanzar Ping al Servidor
    </button>
    
    <div class="pantalla-consola {estado}">
      {mensajeServidor}
    </div>
  </div>
</main>

<style>
  main {
    text-align: center;
    font-family: system-ui, sans-serif;
    padding: 2rem;
    color: #333;
  }
  .panel {
    background: #f8f9fa;
    border: 1px solid #dee2e6;
    padding: 2rem;
    border-radius: 12px;
    max-width: 500px;
    margin: 0 auto;
    box-shadow: 0 4px 6px rgba(0,0,0,0.05);
  }
  button {
    background-color: #0d6efd;
    color: white;
    border: none;
    padding: 12px 24px;
    font-size: 16px;
    font-weight: bold;
    border-radius: 8px;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  button:hover {
    background-color: #0b5ed7;
  }
  .pantalla-consola {
    margin-top: 1.5rem;
    padding: 1rem;
    border-radius: 8px;
    font-family: monospace;
    font-size: 14px;
    background-color: #212529;
    color: white;
  }
  .exito { color: #20c997; border-left: 4px solid #20c997; }
  .error { color: #dc3545; border-left: 4px solid #dc3545; }
  .neutral { color: #adb5bd; border-left: 4px solid #adb5bd; }
</style>