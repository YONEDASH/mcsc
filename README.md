# shadercompat

Mod support for Minecraft shaders.
Feel free to add support for other mods by submitting pull requests.

## Supported Shaders

- [BSL Shaders](https://www.bslshaders.com)

## How it works

This tool works by mapping categorized entries (like blocks and items) to the specific the shader categories.

## Usage

```
 -c string
        file containing categories
  -m string
        directory containing grouped mappings of mods
  -s string
        directory containing shaders
  -shader string
        shader name based on name set in shaders
  -source string
        shader source directory path OR zip file path
```

## Configuration

Blocks, items etc. will be referred to as entries.

### Mod support

A mods blocks, items etc. are categorized in the `mods/` folder inside a *single* `.gm` file (a single file per mod!).
**gm** stands for **g**rouped **m**apping.

#### Categories & Entries

To group entries by a category the first character of the line needs to start with a `[`, followed by the category and
ending with `]`. Make sure to remove trailing whitespace.
Categories themselves are defined in `global/categories.json`.

##### Example

```
[BlockFlatSingle]
minecraft:dandelion
minecraft:poppy
```

#### Namespaces

In order to prevent verbose modname prefixes (like seen above with `minecraft:`) you can use namespaces.
The first character of the line needs to start with `$` followed by the namespace.

##### Example

```
$minecraft:

[BlockFlatSingle]
dandelion
poppy
```

The `dandelion` entry will be mapped as `minecraft:dandelion` and `rose` as `minecraft:rose`.

#### Comments

In order to write a comment the first character of the line needs be to a `#`. The entire line will be ignored.
For example:

```
# Hello World
```

### Shader support

In order to map a specified category's entries to a *specific* shader the mapping in `shaders/` are being used. These
mappings are being specified in the `.json` format. Comments are **not** allowed. Feel free to look at exisiting shader
mappings in the `shaders/` directory.
There is no need to modify a shader's mappings in order to add support for a mod. Mod mappings are supposed to be
independent of shader mappings.

#### Name

`name` specifies the shader's name and is used by the `-shader` flag to look up the correct mapping.

#### Types

`types` specifies the shader's types like blocks, items or entities. A type is structured as follows:

```
"TYPENAME": {
  "file_path": "FILEPATH"
}
```

`TYPENAME` being the type's name like `block` and `FILEPATH` the relative location of the `.properties` file in the
shader's source.
Typically types should look like this:

```
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

`separator` specifies by which character(s) each entry in the `.properties` file specified by the type will be
separated. Typically it is a space (` `).

#### Mappings

`mappings` specifies the mapping from a category (from `global/categories.json`) to a shader category. Those mappings
are structured as follows:

```
"mappings": {
    "TYPE": {
      "CATEGORY": [
            {
            "to": "SHADER_PROPERTIES_KEY"
            }
        ],
    }
}
```

`TYPE` is a type which is specified in `types` like `block`. `CATEGORY` is a category from `global/categories.json`. The
category's value is a list of the locations you want to map to. - That means you can map a entry to multiple different
shader categories. `SHADER_PROPERTIES_KEY` specifies the key of a category in the type's `.properties` file.

##### Transformers

To make mapping two-block plants easy there are transformers. A transformer can be specified by setting a `transformer`
value. Leaving it blank or not assigning it will not use a transformer.
Current transformers are listed below.

- `halfUpper` - Appends `:half=upper` to each entry making the shader target the upper block.
- `halfLower` - Appends `:half=lower` to each entry making the shader target the lower block.
- `halfUpperLower` - Combines `halfUpper` and `halfLower`.
  To get back to the example of two-block plants - that's how you would utilize transformers:

```
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
      ],
    }
}
```
