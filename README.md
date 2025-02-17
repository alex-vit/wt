# wiki-translate
A CLI to translate names or expressions that are hard to translate directly using a dictionary. It works by:
1. finding an article with a matching title
2. getting translations from the language switch menu

## Installation

```sh
go install github.com/alex-vit/wt@latest
```

## Usage examples

```sh
> wt # shows docs
> wt egg salad
en: Egg salad                      https://en.wikipedia.org/wiki/Egg_salad
es: Ensaladilla de huevos          https://es.wikipedia.org/wiki/Ensaladilla_de_huevos
fr: Salade aux œufs                https://fr.wikipedia.org/wiki/Salade_aux_%C5%93ufs
> wt from=pt coelho -save -settings
/Users/alex/Library/Application Support/wt/settings.json:
{
  "target_languages": [
    "en",
    "es",
    "fr",
    "pt"
  ],
  "source_language": "pt"
}
pt: Coelho                         https://pt.wikipedia.org/wiki/Coelho
en: Rabbit                         https://en.wikipedia.org/wiki/Rabbit
es: Conejo                         https://es.wikipedia.org/wiki/Conejo
fr: Lapin                          https://fr.wikipedia.org/wiki/Lapin
```
