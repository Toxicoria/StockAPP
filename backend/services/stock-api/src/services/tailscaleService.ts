export async function generarAuthKeyTailscale(nombreNegocio: string): Promise<string> {
  const apiKey = process.env.TAILSCALE_API_KEY;
  const clientId = process.env.TAILSCALE_CLIENT_ID;
  const clientSecret = process.env.TAILSCALE_CLIENT_SECRET;
  const tailnet = process.env.TAILSCALE_TAILNET || '-';

  // Si hay OAuth Client ID y Secret, obtener primero un Access Token
  let bearerToken = apiKey;

  if (!bearerToken && clientId && clientSecret) {
    try {
      const params = new URLSearchParams({
        client_id: clientId,
        client_secret: clientSecret,
      });
      const tokenResp = await fetch('https://api.tailscale.com/api/v2/oauth/token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: params,
      });
      if (tokenResp.ok) {
        const tokenData = (await tokenResp.json()) as { access_token?: string };
        bearerToken = tokenData.access_token;
      } else {
        console.warn('⚠️ No se pudo obtener Token OAuth de Tailscale:', await tokenResp.text());
      }
    } catch (e) {
      console.warn('⚠️ Error conectando a OAuth de Tailscale:', e);
    }
  }

  // Si tenemos un token válido (apiKey u OAuth token), solicitar la Auth Key reutilizable
  if (bearerToken) {
    try {
      const response = await fetch(`https://api.tailscale.com/api/v2/tailnet/${tailnet}/keys`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${bearerToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          capabilities: {
            devices: {
              create: {
                reusable: true,
                ephemeral: false,
                preauthorized: true,
                tags: ['tag:cliente'],
              },
            },
          },
          expirySeconds: 7776000, // 90 días de validez para registros
          description: `Key reutilizable StockAPP - ${nombreNegocio}`,
        }),
      });

      if (response.ok) {
        const data = (await response.json()) as { key?: string };
        if (data.key) {
          console.log(`✅ Auth Key de Tailscale generada exitosamente para ${nombreNegocio}`);
          return data.key;
        }
      } else {
        console.error('❌ Error de la API de Tailscale:', await response.text());
      }
    } catch (e) {
      console.error('❌ Error llamando a la API de Tailscale:', e);
    }
  }

  // Fallback para desarrollo / test si no hay claves API configuradas
  const fallbackKey = process.env.TS_AUTHKEY || 'tskey-auth-dev-local-demo-key';
  console.log(`ℹ️ Usando Auth Key por defecto (${fallbackKey.substring(0, 15)}...) para ${nombreNegocio}`);
  return fallbackKey;
}
