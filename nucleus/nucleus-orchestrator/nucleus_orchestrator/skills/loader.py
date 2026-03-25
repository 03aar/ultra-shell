"""SkillLoader — discovers and parses YAML skill definitions."""

from __future__ import annotations

import importlib
import os
from pathlib import Path
from typing import Any


def _load_yaml(path: Path) -> dict[str, Any]:
    """Load a YAML file, using PyYAML if available, otherwise a minimal parser."""
    try:
        yaml = importlib.import_module("yaml")
        with open(path, "r", encoding="utf-8") as fh:
            return yaml.safe_load(fh) or {}  # type: ignore[no-any-return]
    except ImportError:
        # Minimal key-value fallback for environments without PyYAML
        data: dict[str, Any] = {}
        with open(path, "r", encoding="utf-8") as fh:
            for line in fh:
                line = line.strip()
                if not line or line.startswith("#"):
                    continue
                if ":" in line:
                    key, _, val = line.partition(":")
                    data[key.strip()] = val.strip()
        return data


class SkillDefinition:
    """Parsed representation of a YAML skill file."""

    def __init__(self, raw: dict[str, Any], source_path: Path) -> None:
        self.name: str = raw.get("name", source_path.stem)
        self.description: str = raw.get("description", "")
        self.version: str = raw.get("version", "1.0.0")
        self.author: str = raw.get("author", "")
        self.tags: list[str] = raw.get("tags", [])
        self.parameters: list[dict[str, Any]] = raw.get("parameters", [])
        self.steps: list[dict[str, Any]] = raw.get("steps", [])
        self.rollback: list[dict[str, Any]] = raw.get("rollback", [])
        self.source_path = source_path
        self._raw = raw

    def __repr__(self) -> str:
        return f"<SkillDefinition {self.name!r} ({len(self.steps)} steps)>"


class SkillLoader:
    """Loads YAML skill files from one or more directories."""

    def __init__(self, directories: list[str | Path] | None = None) -> None:
        self._directories: list[Path] = []
        if directories:
            self._directories = [Path(d) for d in directories]
        else:
            # Default: look for a skills/ dir next to the package and in ~/.nucleus/skills
            default_dirs = [
                Path(__file__).resolve().parent.parent.parent / "skills",
                Path.home() / ".nucleus" / "skills",
            ]
            self._directories = [d for d in default_dirs if d.is_dir()]

        self._skills: dict[str, SkillDefinition] = {}

    def load_all(self) -> dict[str, SkillDefinition]:
        """Scan all directories and load every YAML skill file found."""
        self._skills.clear()
        for directory in self._directories:
            if not directory.is_dir():
                continue
            for entry in sorted(directory.iterdir()):
                if entry.suffix in (".yaml", ".yml") and entry.is_file():
                    self._load_file(entry)
        return dict(self._skills)

    def load_file(self, path: str | Path) -> SkillDefinition:
        """Load a single skill file by path."""
        return self._load_file(Path(path))

    def get(self, name: str) -> SkillDefinition | None:
        """Retrieve a previously loaded skill by name."""
        if not self._skills:
            self.load_all()
        return self._skills.get(name)

    def list_skills(self) -> list[str]:
        """Return sorted names of all loaded skills."""
        if not self._skills:
            self.load_all()
        return sorted(self._skills.keys())

    def _load_file(self, path: Path) -> SkillDefinition:
        raw = _load_yaml(path)
        skill = SkillDefinition(raw, path)
        self._skills[skill.name] = skill
        return skill
