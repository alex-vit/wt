# wiki-translate
A CLI to translate names or expressions that are hard to translate directly using a dictionary. It works by:
1. finding an article with a matching title
2. getting translations from the language switch menu

## Installation

```sh
go install github.com/alex-vit/wt@latest
```

## Usage

```sh
> wt [from=lv] naked mole rat
en: Naked mole-rat                 https://en.wikipedia.org/wiki/Naked_mole-rat
es: Heterocephalus glaber          https://es.wikipedia.org/wiki/Heterocephalus_glaber
fr: Rat-taupe nu                   https://fr.wikipedia.org/wiki/Rat-taupe_nu
lt: Plikasis smėlrausis            https://lt.wikipedia.org/wiki/Plikasis_sm%C4%97lrausis
lv: ???
ru: Голый землекоп                 https://ru.wikipedia.org/wiki/%D0%93%D0%BE%D0%BB%D1%8B%D0%B9_%D0%B7%D0%B5%D0%BC%D0%BB%D0%B5%D0%BA%D0%BE%D0%BF
```

To change the target languages, modify:

```go
targetLangs := []string{"en", "es", "fr", "ru", "lv", "lt"}
```

and recompile.
