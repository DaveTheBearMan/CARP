"""
Written by: David Girard
Date: 10/2/2024
"""

from textual.app import App, ComposeResult
from textual.containers import ScrollableContainer
from textual.widgets import Header, Footer, Button, Static


class ProxyManager(App):
	"""Textual app to manage the manager"""

	def compose(self) -> ComposeResult:
		"""Created child widgets for the app"""
		yield Header()
		yield Footer()

	def 
