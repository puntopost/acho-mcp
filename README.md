# Acho MCP

Acho es un **servidor de memoria persistente para agentes de programación** (Claude Code, OpenCode, Cursor, etc.) vía MCP.

## Qué problema resuelve

Los agentes pierden todo el contexto entre sesiones y compactaciones: reglas que aplicabas se olvidan, decisiones de arquitectura se reabren, bugs resueltos vuelven, convenciones se saltan.

Acho guarda ese contexto en local, lo inyecta al arrancar cada sesión, y lo pone al alcance del agente por SQL estructurado.

## Idea en una frase

El usuario define:
1. **Reglas** (texto libre que se inyecta como `==MANDATORY==` al arrancar cada sesión).
2. **Tipos de registro** con su JSON Schema (qué puede guardarse y cómo).

El agente guarda registros (JSON que cumple el schema del tipo) y los consulta con SQL cuando los necesita.

Todo lo que guardes puede ser **global** (visible en cualquier proyecto) o **del proyecto actual** (scopeado al directorio de trabajo).

## Instalación

1. Descarga el zip de la última release desde [GitHub Releases](https://github.com/puntopost/acho-mcp/releases) (`acho-vX.Y.Z-linux-amd64.zip`).
2. Descomprime y coloca el binario `acho` en tu `PATH` (por ejemplo `~/.local/bin`).
3. Regístralo en tu agente:

```
acho agent-setup claude       # Claude Code
acho agent-setup opencode     # OpenCode
```

Eso instala el plugin que arranca Acho como servidor MCP, un hook que carga las reglas al empezar cada sesión, y las instrucciones que le dicen al agente cómo usarlo.

## Activación por proyecto

Acho viene **desactivado por defecto** en cada proyecto. Hasta que lo actives explícitamente, el servidor MCP arranca sin exponer ninguna tool en ese proyecto (para no contaminar con memoria ajena un repo donde aún no quieres usarlo).

```
acho project enable          # activa Acho en el proyecto actual
acho project disable         # lo desactiva de nuevo
acho project status          # muestra el estado y el nombre detectado
```

Tras `enable` o `disable`, reconecta el MCP en tu agente para que aplique el cambio.

## Cómo funciona

Acho gestiona tres conceptos: **reglas**, **tipos** y **registros**. Los tres pueden ser globales (visibles en todos tus proyectos) o del proyecto actual (scopeados al directorio de trabajo).

### Reglas

Son las órdenes que el usuario le dicta al modelo y que quiere que se cumplan siempre. Deben ser sencillas, directas y sin ambigüedades, lo más concisas posibles. Acho las inyecta en cada arranque de sesión, así que el modelo las trata como instrucciones obligatorias a lo largo de la conversación.

### Tipos y registros

Los tipos y los registros son conceptos intrínsecamente relacionados. Un **tipo** es la definición dinámica del esquema de un registro; un **registro** es una entrada concreta que cumple ese esquema.

Por ejemplo, puedes crear un tipo `animal` con un esquema que guarde `nombre`, `especie` y `familia`. A partir de ese momento, si le dices al modelo *"guárdame un tipo de animal nuevo, los perros"*, el modelo creará un registro de tipo `animal`, inferirá los campos del esquema que tenga claros y preguntará al usuario por los que no.

Esto lo hace la inteligencia del propio modelo: Acho solo define los esquemas JSON y se encarga de validar que los registros los cumplan. Cómo se rellenan y cuándo se consultan lo decide el modelo interpretando la conversación.

### Sin límites

El usuario puede crear el conjunto de reglas, tipos y registros que quiera para resolver sus problemas cotidianos de la forma que mejor le parezca. Mediante órdenes en lenguaje natural, el modelo creará, editará, borrará o listará cualquier regla, tipo o registro; y sabrá presentar los resultados del modo que se le pida.

### Borrado seguro

El borrado es *soft* por defecto: las entradas desaparecen de las consultas pero se pueden recuperar si se borraron por error. La CLI ofrece un comando específico (`acho purge`) para eliminarlas de forma definitiva y controlada cuando toque.

## Ejemplos de uso

Las líneas en cursiva son lo que le dices tú al agente; debajo, lo que el agente termina haciendo.

### 1. Guardar una regla

> *"Guárdate como regla global que siempre que toques código Go corras el linter antes de darlo por terminado."*

El agente crea una regla con `project:"global"`. En la siguiente sesión, al arrancar, esa regla llega envuelta en `==MANDATORY==` y queda cargada para toda la conversación.

### 2. Crear un tipo y guardar registros con él

> *"Voy a empezar a apuntar los bugs que resolvemos. Crea un tipo `bugfix` con campos `problema`, `causa` y `solucion`, todos obligatorios."*

El agente define el tipo con su JSON Schema. A partir de ahí, cuando le digas *"apúntate este bug"*, creará un registro validado contra ese schema. Si intenta guardar algo que no cumple el schema, la validación falla y el agente lo corrige.

### 3. Combinar reglas + tipos para contexto automático

La combinación potente: una regla que le dice al agente *"antes de hacer X, consulta los registros de tipo Y"*.

Por ejemplo, defines un tipo `decision` (con campos `tema`, `eleccion`, `motivo`) y una regla global *"antes de proponer arquitectura, consulta los registros de tipo `decision` del proyecto"*. A partir de entonces, cada vez que discutas arquitectura, el agente tirará de su memoria antes de hablar.

Así Acho deja de ser un cajón de notas y se convierte en un contexto vivo que el propio agente sabe cuándo leer.

## CLI

Para inspección y gestión sin pasar por el agente:

```
acho rules list
acho rules delete <id>
acho types list
acho types delete <name> [--force]
acho registries list [flags]
acho registries get <id>
acho registries delete <id>
acho stats
acho export [file]
acho import [file]
acho purge                                # elimina soft-deleted definitivamente
acho project [status|enable|disable|rename]
acho config
acho config show
acho agent-setup [claude|opencode]
```

Para **guardar, actualizar o buscar registros** usa el agente vía MCP — el CLI no incluye esos comandos a propósito (la idea es que lo escriba el agente).

Más detalle con `acho --help` y `acho <comando> --help`.

## Configuración

- Base de datos: `~/.acho/acho.db`.
- Configuración: `~/.acho/config.json`.
- Logs: `~/.acho/acho.log`.

Puedes mover toda la estructura con la variable de entorno `ACHO_PATH`.
