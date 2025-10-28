# Tribal Framework

This repository uses the Tribal framework for graph-based development.

## Overview

Tribal is a framework for organizing and managing development work through graph structures.
Each graph represents a specific feature, component, or concept within the codebase.

## Usage

- Use 'tribal checkout -g"<graph title>"' to create or retrieve a graph
- Use 'tribal search --context <description>' to find related graphs
- Use 'tribal add -A' to stage your graph changes
- Use 'tribal commit -m"<message>"' to commit your graph
- Use 'tribal push' to push changes to the remote

## Graph Structure

Graphs consist of nodes and edges that represent relationships between code components,
features, or concepts. Each node can contain markup for documentation and context.
