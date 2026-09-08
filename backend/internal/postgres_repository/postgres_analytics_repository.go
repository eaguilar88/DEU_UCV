package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/eaguilar88/deu/internal/group_analytics"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
)

func (r *PostgresRepository) GetActivityParticipantsMetrics(ctx context.Context, year int, groupID string) ([]group_analytics.ActivityParticipantsMetric, error) {
	sqlQuery, args, err := queries.GetActivityParticipantsMetricsQuery(year, groupID).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []group_analytics.ActivityParticipantsMetric
	for rows.Next() {
		var item group_analytics.ActivityParticipantsMetric
		if err := rows.Scan(&item.Lugar, &item.CantidadEsperada, &item.CantidadReal, &item.Integrantes); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) GetActivitiesByStateMetrics(ctx context.Context, year int, groupID string) ([]group_analytics.ActivitiesByStateMetric, error) {
	sqlQuery, args, err := queries.GetActivitiesByStateMetricsQuery(year, groupID).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []group_analytics.ActivitiesByStateMetric
	for rows.Next() {
		var item group_analytics.ActivitiesByStateMetric
		if err := rows.Scan(&item.Lugar, &item.CantidadReal); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) GetGroupParticipantsMetrics(ctx context.Context, year int, groupID string) ([]group_analytics.GroupParticipantsMetric, error) {
	sqlQuery, args, err := queries.GetGroupParticipantsMetricsQuery(year, groupID).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []group_analytics.GroupParticipantsMetric
	for rows.Next() {
		var item group_analytics.GroupParticipantsMetric
		if err := rows.Scan(&item.Lugar, &item.CantidadEsperada, &item.CantidadReal, &item.Integrantes, &item.CantidadDeVeces); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) GetActivitiesByCityMetrics(ctx context.Context, year int, groupID string) ([]group_analytics.ActivitiesByCityMetric, error) {
	sqlQuery, args, err := queries.GetActivitiesByCityMetricsQuery(year, groupID).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []group_analytics.ActivitiesByCityMetric
	for rows.Next() {
		var item group_analytics.ActivitiesByCityMetric
		if err := rows.Scan(&item.Lugar, &item.CantidadReal); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) GetYearlyActivitiesMetrics(ctx context.Context, startYear, endYear int, groupID string) ([]group_analytics.YearlyActivitiesMetric, error) {
	sqlQuery, args, err := queries.GetYearlyActivitiesMetricsQuery(startYear, endYear, groupID).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []group_analytics.YearlyActivitiesMetric
	for rows.Next() {
		var item group_analytics.YearlyActivitiesMetric
		if err := rows.Scan(&item.Lugar, &item.CantidadReal); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) GetKnowledgeAreasByYearMetrics(ctx context.Context, startYear, endYear int, masterAreas []string, groupID string) ([]map[string]interface{}, error) {
	sqlQuery, args, err := queries.GetRawKnowledgeAreasByYearQuery(startYear, endYear, groupID).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Crear un set con las áreas maestras en mayúsculas para comparación rápida
	masterMap := make(map[string]string)
	for _, area := range masterAreas {
		masterMap[strings.ToUpper(strings.TrimSpace(area))] = area
	}

	// Inicializar mapa de acumulación por año
	yearMap := make(map[string]map[string]int)
	for y := startYear; y <= endYear; y++ {
		yearStr := fmt.Sprintf("%d", y)
		yearMap[yearStr] = make(map[string]int)
		for _, area := range masterAreas {
			yearMap[yearStr][area] = 0
		}
	}

	for rows.Next() {
		var yearStr string
		var rawAreas pq.StringArray

		if err := rows.Scan(&yearStr, &rawAreas); err != nil {
			return nil, err
		}

		if _, exists := yearMap[yearStr]; !exists {
			continue
		}

		for _, area := range rawAreas {
			cleanArea := strings.ToUpper(strings.TrimSpace(area))
			if cleanArea == "" {
				continue
			}

			if originalName, found := masterMap[cleanArea]; found {
				yearMap[yearStr][originalName]++
			} else {
				yearMap[yearStr]["Otros"]++
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var response []map[string]interface{}
	for y := startYear; y <= endYear; y++ {
		yearStr := fmt.Sprintf("%d", y)
		item := make(map[string]interface{})
		item["lugar"] = yearStr
		for k, v := range yearMap[yearStr] {
			item[k] = v
		}
		response = append(response, item)
	}

	return response, nil
}
