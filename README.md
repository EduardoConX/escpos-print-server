# Servidor de impresión ESC/POS

Servidor HTTP local para enviar tickets ESC/POS a una impresora térmica instalada en Windows. Está preparado para recibir solicitudes del POS por la red y enviar los datos directamente a la cola de impresión de Windows.

Basado en el proyecto de [Parzibyte](https://github.com/parzibyte/ejemplos-plugin-impresoras-termicas-v2).

## Requisitos

- Windows en la computadora donde está instalada la impresora.
- Una impresora compatible con ESC/POS instalada y visible en Windows.
- Go instalado en la computadora donde se compilará el servidor. Go no hace falta en la computadora de la impresora si se copia y ejecuta el `.exe` compilado.

La cola local de esta instalación se llama `XP-58`. `TicketsPrinter` es el nombre compartido de Windows; los equipos POS envían solicitudes al servidor y este imprime por la cola local `XP-58`.

## Compilar

Desde PowerShell, en la carpeta del proyecto:

```powershell
go build -o escpos-print-server.exe .
```

Se puede copiar `escpos-print-server.exe` a la computadora de la impresora. La impresora `XP-58` debe estar instalada allí con ese nombre.

## Configuración y arranque

El servidor usa estas variables de entorno:

| Variable | Predeterminado | Descripción |
| --- | --- | --- |
| `PORT` | `8080` | Puerto HTTP del servidor. |
| `PRINTER_NAME` | `XP-58` | Nombre de la cola de impresión de Windows. |
| `LOG_LEVEL` | `info` | Nivel indicado al logger. |

Ejemplo de inicio en PowerShell:

```powershell
$env:PORT="8080"
$env:PRINTER_NAME="XP-58"
$env:LOG_LEVEL="info"
.\escpos-print-server.exe
```

Para ejecutarlo desde el código fuente, usar `go run .` en lugar del `.exe`. La terminal debe permanecer abierta mientras el servidor esté ejecutándose. Para detenerlo, presiona `Ctrl+C`.

## Probar una impresión local

Con el servidor activo, abre otra ventana de PowerShell y envía un ticket de prueba:

```powershell
$ticket = @{
	operations = @(
		@{ action = "alignment"; data = "C" }
		@{ action = "boldText"; data = "1" }
		@{ action = "text"; data = "PRUEBA ESC/POS" }
		@{ action = "enter"; data = "" }
		@{ action = "boldText"; data = "0" }
		@{ action = "text"; data = "Conexion correcta" }
		@{ action = "feed"; data = "2" }
	)
} | ConvertTo-Json -Depth 5

Invoke-RestMethod `
	-Uri "http://localhost:8080/" `
	-Method Post `
	-ContentType "application/json" `
	-Body $ticket
```

La respuesta correcta es `status: ok`. También revisa que el ticket aparezca en la impresora y en la cola de impresión de Windows. Si hay un error, consulta los mensajes de la consola donde está corriendo el servidor.

## Conectar el POS desde otra computadora

Obtén la dirección IP local de la computadora que ejecuta el servidor. Desde el POS, envía el mismo `POST` a:

```text
http://IP-DE-LA-PC:8080/
```

Por ejemplo, si la PC del servidor tiene la IP `192.168.1.25`, la URL será `http://192.168.1.25:8080/`. Si no hay conexión, permite el puerto configurado en el Firewall de Windows para la red privada.

El campo `printer` del JSON es opcional. Si se omite, el servidor utiliza `PRINTER_NAME`. Si se proporciona, ese valor debe coincidir con el nombre de una cola disponible para el proceso del servidor.

## Formato de la solicitud

```json
{
	"operations": [
		{ "action": "alignment", "data": "C" },
		{ "action": "boldText", "data": "1" },
		{ "action": "text", "data": "Mi tienda" },
		{ "action": "enter", "data": "" },
		{ "action": "boldText", "data": "0" },
		{ "action": "separator", "data": "-" },
		{ "action": "text", "data": "Gracias por su compra" },
		{ "action": "feed", "data": "2" },
		{ "action": "cut", "data": "" }
	]
}
```

Acciones disponibles: `text`, `enter`, `line`, `separator`, `alignment` (`L`, `C`, `R`), `boldText` (`1` para activar, `0` para desactivar), `fontSize` (`ancho,alto`), `feed` (cantidad de líneas), `cut` y `openCashDrawer`.

La respuesta exitosa es HTTP `200` con `{"status":"ok"}`. Las solicitudes inválidas responden `400`; los errores al enviar a la impresora responden `500`.

## Seguridad y operación

- El servidor escucha en el puerto configurado en las interfaces de red de la PC. No lo expongas directamente a Internet.
- Restringe el acceso al puerto en el Firewall a la red privada de la computadora; el servicio actualmente no requiere autenticación.
- Para operación continua, configura el `.exe` para iniciar con Windows mediante el Programador de tareas o un servicio de Windows. El proyecto no instala ese servicio automáticamente.
- Las acciones de corte y apertura del cajón dependen de que el modelo y su configuración ESC/POS las soporten.