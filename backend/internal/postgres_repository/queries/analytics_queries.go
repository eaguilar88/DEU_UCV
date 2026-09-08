package queries

import (
	sq "github.com/Masterminds/squirrel"
)

// 1. Participantes por Actividad
func GetActivityParticipantsMetricsQuery(year int, groupID string) sq.SelectBuilder {
	q := psql.Select(
		"a.name AS lugar",
		"COALESCE(a.stimated_participants, 0) AS cantidad_esperada",
		"COALESCE(a.actual_participants, 0) AS cantidad_real",
		"COALESCE(a.group_participants, 0) AS integrantes",
	).
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.deleted_at": nil}).
		Where("EXTRACT(YEAR FROM a.date_start) = ?", year).
		Where("(COALESCE(a.stimated_participants, 0) > 0 OR COALESCE(a.actual_participants, 0) > 0)")

	if groupID != "" {
		q = q.Where(sq.Eq{"a.group_id": groupID})
	}
	return q
}

// 2. Volumen por Estado
func GetActivitiesByStateMetricsQuery(year int, groupID string) sq.SelectBuilder {
	q := psql.Select(
		"TRIM(SPLIT_PART(a.location, ',', 2)) AS lugar",
		"COUNT(*) AS cantidad_real",
	).
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.deleted_at": nil}).
		Where("EXTRACT(YEAR FROM a.date_start) = ?", year).
		Where("TRIM(SPLIT_PART(a.location, ',', 2)) <> ''")

	if groupID != "" {
		q = q.Where(sq.Eq{"a.group_id": groupID})
	}
	return q.GroupBy("TRIM(SPLIT_PART(a.location, ',', 2))")
}

// 3. Participantes por Grupo
func GetGroupParticipantsMetricsQuery(year int, groupID string) sq.SelectBuilder {
	q := psql.Select(
		"g.name AS lugar",
		"SUM(COALESCE(a.stimated_participants, 0)) AS cantidad_esperada",
		"SUM(COALESCE(a.actual_participants, 0)) AS cantidad_real",
		"SUM(COALESCE(a.group_participants, 0)) AS integrantes",
		"COUNT(a.id) AS cantidad_de_veces",
	).
		From(activitiesTableName + " AS a").
		Join(groupsTableName + " AS g ON a.group_id = g.id").
		Where(sq.Eq{"a.deleted_at": nil}).
		Where("EXTRACT(YEAR FROM a.date_start) = ?", year).
		Where("(COALESCE(a.stimated_participants, 0) > 0 OR COALESCE(a.actual_participants, 0) > 0)")

	if groupID != "" {
		q = q.Where(sq.Eq{"a.group_id": groupID})
	}
	return q.GroupBy("g.name")
}

// 4. Actividades por Ciudad/Municipio
func GetActivitiesByCityMetricsQuery(year int, groupID string) sq.SelectBuilder {
	q := psql.Select(
		"TRIM(SPLIT_PART(a.location, ',', 3)) AS lugar",
		"COUNT(*) AS cantidad_real",
	).
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.deleted_at": nil}).
		Where("EXTRACT(YEAR FROM a.date_start) = ?", year).
		Where("TRIM(SPLIT_PART(a.location, ',', 3)) <> ''")

	if groupID != "" {
		q = q.Where(sq.Eq{"a.group_id": groupID})
	}
	return q.GroupBy("TRIM(SPLIT_PART(a.location, ',', 3))")
}

// 5. Histórico por Año
func GetYearlyActivitiesMetricsQuery(startYear, endYear int, groupID string) sq.SelectBuilder {
	q := psql.Select(
		"CAST(EXTRACT(YEAR FROM a.date_end) AS TEXT) AS lugar",
		"COUNT(*) AS cantidad_real",
	).
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.deleted_at": nil}).
		Where("EXTRACT(YEAR FROM a.date_end) BETWEEN ? AND ?", startYear, endYear)

	if groupID != "" {
		q = q.Where(sq.Eq{"a.group_id": groupID})
	}
	return q.GroupBy("EXTRACT(YEAR FROM a.date_end)").OrderBy("lugar ASC")
}

// 6. Áreas por Año (Obtiene año y un array de áreas de conocimiento por actividad)
func GetRawKnowledgeAreasByYearQuery(startYear, endYear int, groupID string) sq.SelectBuilder {
	q := psql.Select(
		"CAST(EXTRACT(YEAR FROM a.date_start) AS TEXT) AS anio",
		"a.knowledge_area",
	).
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.deleted_at": nil}).
		Where("EXTRACT(YEAR FROM a.date_start) BETWEEN ? AND ?", startYear, endYear)

	if groupID != "" {
		q = q.Where(sq.Eq{"a.group_id": groupID})
	}
	return q
}
