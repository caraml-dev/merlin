from docutils import nodes
from ipypublish.sphinx.notebook.nodes import AdmonitionNode
from ipypublish.sphinx.notebook.nodes import CodeAreaNode
from ipypublish.sphinx.notebook.nodes import FancyOutputNode

def visit_AdmonitionNode(self, node):
    if 'warning' in node['classes']:
        self.visit_warning(node)
    else:
        self.visit_note(node)

def depart_AdmonitionNode(self, node):
    if 'warning' in node['classes']:
        self.depart_warning(node)
    else:
        self.depart_note(node)

def visit_CodeAreaNode(self, node):
    pass

def depart_CodeAreaNode(self, node):
    pass

def visit_FancyOutputNode(self, node):
    raise nodes.SkipNode

def setup(app):
    app.require_sphinx('1.0')

    app.add_node(AdmonitionNode, confluence=
        (visit_AdmonitionNode, depart_AdmonitionNode), override=True)
    app.add_node(CodeAreaNode, confluence=
        (visit_CodeAreaNode, depart_CodeAreaNode), override=True)
    app.add_node(FancyOutputNode, confluence=
        (visit_FancyOutputNode, None), override=True)