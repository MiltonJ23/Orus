package port

import "context"

// ContentReader est le port pour extraire le contenu textuel lisible d'un fichier
type ContentReader interface {
	// ReadBookText extrait tout le texte d'un livre et le découpe en chunks de taille raisonnable.
	// Chaque chunk représente une "page" dans le lecteur.
	ReadBookText(ctx context.Context, filePath string) ([]string, error)
}
