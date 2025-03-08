# ShaderCompat

Mod support for Minecraft shaders.
Feel free to add support for a mod or shader by submitting a pull request.

## Supported Shaders

- [BSL Shaders](https://www.bslshaders.com)

## How It Works

This tool works by mapping categorized entries (such as blocks and items) to specific shader categories.

## Usage

```
  -c string
        File containing categories.
  -m string
        Directory containing grouped mappings of mods.
  -s string
        Directory containing shaders.
  -shader string
        Shader name based on the name set in shaders.
  -source string
        Shader source directory path or zip file path.
```

## Configuration

Blocks, items, etc., will be referred to as entries.

### Mod Support

A mod's entries are categorized in the `mods/` folder inside a *single* `.gm` file (one file per mod!).
**gm** stands for **g**rouped **m**apping.

#### Categories & Entries

To group entries by a category, the first character of the line must start with `[`, followed by the category name, and ending with `]`. Make sure to remove trailing whitespace.
Categories themselves are defined in `global/categories.json`.

##### Example

```
[BlockFlatSingle]
minecraft:dandelion
minecraft:poppy
```

#### Namespaces

To prevent verbose mod name prefixes (such as `minecraft:`), you can use namespaces.
The first character of the line must start with `$`, followed by the namespace.

##### Example

```
$minecraft:

[BlockFlatSingle]
dandelion
poppy
```

The `dandelion` entry will be mapped as `minecraft:dandelion`, and `poppy` as `minecraft:poppy`.

#### Comments

To write a comment, the first character of the line must be `#`. The entire line will be ignored.
For example:

```
# Hello World
```

### Shader Support

To map a specified category's entries to a *specific* shader, the mappings in `shaders/` are used. These mappings are specified in `.json` format. Comments are **not** allowed. Feel free to look at existing shader mappings in the `shaders/` directory.
There is no need to modify a shader's mappings to add support for a mod. Mod mappings are independent of shader mappings.

#### Name

`name` specifies the shader's name and is used by the `-shader` flag to look up the correct mapping.

#### Types

`types` specifies the shader's types, such as blocks, items, or entities. A type is structured as follows:

```json
"TYPENAME": {
    "file_path": "FILEPATH"
}
```

- `TYPENAME`: The type's name, such as `block`.
- `FILEPATH`: The relative location of the `.properties` file in the shader's source.

Typically, types should look like this:

```json
"types": {
    "block": {
        "file_path": "./shaders/block.properties"
    },
    "item": {
        "file_path": "./shaders/item.properties"
    }
}
```

#### Separator

`separator` specifies the character(s) by which each entry in the `.properties` file specified by the type will be separated. Typically, it is a space (` `).

#### Mappings

`mappings` specifies the mapping from a category (from `global/categories.json`) to a shader category. These mappings are structured as follows:

```json
"mappings": {
    "TYPE": {
        "CATEGORY": [
            {
                "to": "SHADER_PROPERTIES_KEY"
            }
        ]
    }
}
```

- `TYPE`: A type specified in `types`, such as `block`.
- `CATEGORY`: A category from `global/categories.json`.
- `SHADER_PROPERTIES_KEY`: Specifies the key of a category in the type's `.properties` file.

The category's value is a list of locations you want to map to, meaning you can map an entry to multiple different shader categories.

##### Transformers

To make mapping two-block plants easier, transformers are available. A transformer can be specified by setting a `transformer` value. Leaving it blank or omitting it means no transformer is used.

Current transformers:

- `halfUpper` - Appends `:half=upper` to each entry, making the shader target the upper block.
- `halfLower` - Appends `:half=lower` to each entry, making the shader target the lower block.
- `halfUpperLower` - Combines `halfUpper` and `halfLower`.

To illustrate how to use transformers for two-block plants:

```json
"mappings": {
    "block": {
        "BlockFlatDouble": [
            {
                "to": "SHADER_PROPERTIES_LOWER_BLOCK_KEY",
                "transformer": "halfLower"
            },
            {
                "to": "SHADER_PROPERTIES_UPPER_BLOCK_KEY",
                "transformer": "halfUpper"
            }
        ]
    }
}
```

