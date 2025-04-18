# Monitor de Contenedores Docker

Este es un monitor de contenedores Docker que envía notificaciones por Telegram cuando detecta que algún contenedor se ha detenido o está en un estado anormal.

## Características

- Monitoreo continuo de contenedores Docker
- Notificaciones por Telegram
- Heartbeats periódicos
- Auto-recuperación mediante supervisor
- Logs detallados
- Contenerizado para fácil despliegue

## Requisitos

- Docker
- Docker Compose
- Token de bot de Telegram
- Chat ID de Telegram

## Configuración

1. Crea un bot en Telegram:
   - Habla con @BotFather
   - Usa el comando `/newbot`
   - Sigue las instrucciones para nombrar tu bot
   - Guarda el token que te proporciona

2. Obtén tu Chat ID:
   - Inicia una conversación con tu bot
   - Visita: `https://api.telegram.org/bot<YourBOTToken>/getUpdates`
   - Busca el "chat_id" en la respuesta JSON

3. Configura las variables de entorno:
   - Edita el archivo `.env` con tu token y chat ID

## Despliegue

1. Construye y ejecuta los contenedores:
```bash
docker-compose up -d
```

2. Verifica los logs:
```bash
docker-compose logs -f
```

## Monitoreo del Monitor

Para verificar el estado del monitor:

```bash
# Ver estado de los contenedores
docker ps -a | grep docker-monitor

# Ver logs del monitor
docker logs docker-monitor

# Ver logs del supervisor
docker logs docker-monitor-supervisor

# Ver estadísticas de recursos
docker stats docker-monitor docker-monitor-supervisor
```

## Notas

- El monitor verifica los contenedores cada 5 minutos
- Envía un heartbeat cada 30 minutos
- El supervisor verifica el estado del monitor cada 5 minutos
- Los logs se almacenan en un volumen Docker #   d o c k e r _ m o n i t o r _ g o  
 