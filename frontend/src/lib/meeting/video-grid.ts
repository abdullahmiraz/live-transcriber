/** Tailwind grid classes for the in-call video stage based on participant count. */
export function videoGridClassForCount(count: number): string {
	if (count <= 1) return 'grid-cols-1 grid-rows-1';
	if (count === 2) return 'grid-cols-1 grid-rows-2 sm:grid-cols-2 sm:grid-rows-1';
	if (count <= 4) return 'grid-cols-2 grid-rows-2';
	if (count <= 6) return 'grid-cols-2 grid-rows-3 lg:grid-cols-3 lg:grid-rows-2';
	return 'grid-cols-3 grid-rows-3';
}
