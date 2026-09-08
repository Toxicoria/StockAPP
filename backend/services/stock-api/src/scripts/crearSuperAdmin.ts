import readline from 'node:readline';
import bcrypt from 'bcryptjs';
import { pool } from '../db/pool.js';

function pedirEntrada(pregunta: string): Promise<string> {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(pregunta, (respuesta) => {
      rl.close();
      resolve(respuesta.trim());
    });
  });
}

async function main() {
  console.log('\n======================================================');
  console.log('   🛡️  StockAPP - Creación de Super Administrador     ');
  console.log('======================================================\n');

  try {
    // Parámetros por línea de comandos o interactivos
    let usuario = process.argv[2];
    let password = process.argv[3];
    let nombre = process.argv[4];
    let email = process.argv[5];

    if (!usuario) {
      usuario = await pedirEntrada('👤 Usuario administrador (ej: admin): ');
    }
    if (!password) {
      password = await pedirEntrada('🔑 Contraseña: ');
    }
    if (!nombre) {
      nombre = await pedirEntrada('🏷️  Nombre completo (ej: Administrador Principal): ');
    }
    if (!email) {
      email = await pedirEntrada('📧 Correo electrónico (opcional): ');
    }

    if (!usuario || !password || !nombre) {
      console.error('\n❌ Error: El usuario, contraseña y nombre son obligatorios.');
      process.exit(1);
    }

    if (password.length < 6) {
      console.error('\n❌ Error: La contraseña debe tener al menos 6 caracteres.');
      process.exit(1);
    }

    // Comprobar si ya existe
    const existe = await pool.query(
      'SELECT id_admin FROM super_admins WHERE LOWER(usuario) = LOWER($1)',
      [usuario],
    );

    if (existe.rowCount && existe.rowCount > 0) {
      console.error(`\n❌ Error: El usuario "${usuario}" ya existe en la base de datos.`);
      process.exit(1);
    }

    const saltRounds = 10;
    const hash = await bcrypt.hash(password, saltRounds);

    const res = await pool.query(
      `INSERT INTO super_admins (usuario, password_hash, nombre, email, totp_activado)
       VALUES ($1, $2, $3, $4, FALSE)
       RETURNING id_admin, usuario, nombre, email`,
      [usuario, hash, nombre, email || null],
    );

    const nuevoAdmin = res.rows[0];

    console.log('\n✅ ¡Super Administrador creado exitosamente!');
    console.log('------------------------------------------------------');
    console.log(` ID:      ${nuevoAdmin.id_admin}`);
    console.log(` Usuario: ${nuevoAdmin.usuario}`);
    console.log(` Nombre:  ${nuevoAdmin.nombre}`);
    console.log(` Email:   ${nuevoAdmin.email || '(sin email)'}`);
    console.log('------------------------------------------------------');
    console.log('ℹ️  En el primer inicio de sesión se le pedirá escanear');
    console.log('   el código QR para configurar la autenticación 2FA.\n');

  } catch (err: any) {
    console.error('\n❌ Error al crear el SuperAdmin:', err.message || err);
  } finally {
    await pool.end();
  }
}

main();
