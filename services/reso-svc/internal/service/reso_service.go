package service

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/reso-svc/internal/domain"
)

type ResoService struct {
	repo domain.NetworkRepository
	log  *zap.Logger
}

func NewResoService(repo domain.NetworkRepository, log *zap.Logger) *ResoService {
	return &ResoService{repo: repo, log: log}
}

func (s *ResoService) AnalyzeNetwork() (*domain.NetworkAnalysisResult, error) {
	persons, err := s.repo.GetAllPersons()
	if err != nil {
		return nil, err
	}

	gangs, err := s.repo.GetAllGangs()
	if err != nil {
		return nil, err
	}

	associations, err := s.repo.GetPersonAssociations()
	if err != nil {
		return nil, err
	}

	adj := make(map[uuid.UUID]map[uuid.UUID]float64)
	for _, p := range persons {
		adj[p.SnisidID] = make(map[uuid.UUID]float64)
	}

	for _, a := range associations {
		weight := a.Confidence
		if adj[a.PersonID1] == nil {
			adj[a.PersonID1] = make(map[uuid.UUID]float64)
		}
		if adj[a.PersonID2] == nil {
			adj[a.PersonID2] = make(map[uuid.UUID]float64)
		}
		adj[a.PersonID1][a.PersonID2] = weight
		adj[a.PersonID2][a.PersonID1] = weight
	}

	communities := labelPropagation(adj)
	modularity := computeModularity(adj, communities)

	networks := make([]domain.CriminalNetwork, 0, len(communities))
	for i, comm := range communities {
		name := "Community-" + time.Now().Format("20060102") + "-" + string(rune('A'+i))
		network := &domain.CriminalNetwork{
			NetworkID:       uuid.New(),
			NetworkName:     &name,
			MemberIDs:       comm.Members,
			CommunitySize:   comm.Size,
			ModularityScore: &modularity,
			DetectedAt:      time.Now(),
			AnalysisVersion: "1.0",
		}
		if err := s.repo.SaveNetwork(network); err != nil {
			s.log.Error("failed to save network", zap.Error(err))
			continue
		}
		networks = append(networks, *network)
	}

	result := &domain.NetworkAnalysisResult{
		Networks:     networks,
		TotalPersons: len(persons),
		TotalGangs:   len(gangs),
		TotalEdges:   len(associations),
		Communities:  len(communities),
		AnalyzedAt:   time.Now(),
	}

	return result, nil
}

func (s *ResoService) GetPersonNetwork(personID uuid.UUID) (map[string]interface{}, error) {
	associations, err := s.repo.GetPersonNetwork(personID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"person_id":    personID,
		"associations": associations,
		"direct_links": len(associations),
	}

	return result, nil
}

func (s *ResoService) GetKeyActors(limit int) ([]domain.KeyActor, error) {
	persons, err := s.repo.GetAllPersons()
	if err != nil {
		return nil, err
	}

	associations, err := s.repo.GetPersonAssociations()
	if err != nil {
		return nil, err
	}

	memberships, err := s.repo.GetPersonGangMemberships()
	if err != nil {
		return nil, err
	}

	adj := make(map[uuid.UUID]map[uuid.UUID]float64)
	for _, p := range persons {
		adj[p.SnisidID] = make(map[uuid.UUID]float64)
	}
	for _, a := range associations {
		if adj[a.PersonID1] == nil {
			adj[a.PersonID1] = make(map[uuid.UUID]float64)
		}
		if adj[a.PersonID2] == nil {
			adj[a.PersonID2] = make(map[uuid.UUID]float64)
		}
		adj[a.PersonID1][a.PersonID2] = a.Confidence
		adj[a.PersonID2][a.PersonID1] = a.Confidence
	}

	pagerank := computePageRank(adj)
	betweenness := computeBetweenness(adj)
	degreeScores := computeDegreeCentrality(adj)

	personMap := make(map[uuid.UUID]domain.Person)
	for _, p := range persons {
		personMap[p.SnisidID] = p
	}

	gangMembership := make(map[uuid.UUID]uuid.UUID)
	for _, m := range memberships {
		gangMembership[m.PersonID] = m.GangID
	}

	actors := make([]domain.KeyActor, 0, len(persons))
	for id, pr := range pagerank {
		bc := betweenness[id]
		dc := degreeScores[id]
		composite := 0.4*pr + 0.4*bc + 0.2*dc

		actor := domain.KeyActor{
			PersonID:         id,
			PageRankScore:    pr,
			BetweennessScore: bc,
			DegreeScore:      dc,
			CompositeScore:   composite,
		}

		if p, ok := personMap[id]; ok {
			actor.Name = p.Name
		}
		if gid, ok := gangMembership[id]; ok {
			actor.GangID = &gid
		}

		actors = append(actors, actor)
	}

	sort.Slice(actors, func(i, j int) bool {
		return actors[i].CompositeScore > actors[j].CompositeScore
	})

	if limit > 0 && len(actors) > limit {
		actors = actors[:limit]
	}

	return actors, nil
}

func (s *ResoService) FindShortestPath(fromID, toID uuid.UUID) (*domain.ShortestPathResult, error) {
	associations, err := s.repo.GetPersonAssociations()
	if err != nil {
		return nil, err
	}

	adj := make(map[uuid.UUID][]uuid.UUID)
	for _, a := range associations {
		adj[a.PersonID1] = append(adj[a.PersonID1], a.PersonID2)
		adj[a.PersonID2] = append(adj[a.PersonID2], a.PersonID1)
	}

	visited := make(map[uuid.UUID]bool)
	queue := [][]uuid.UUID{{fromID}}
	visited[fromID] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		current := path[len(path)-1]

		if current == toID {
			result := &domain.ShortestPathResult{
				FromID:     fromID,
				ToID:       toID,
				Path:       path,
				PathLength: len(path) - 1,
				Exists:     true,
			}
			return result, nil
		}

		for _, neighbor := range adj[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				newPath := make([]uuid.UUID, len(path)+1)
				copy(newPath, path)
				newPath[len(path)] = neighbor
				queue = append(queue, newPath)
			}
		}
	}

	return &domain.ShortestPathResult{
		FromID: fromID,
		ToID:   toID,
		Path:   []uuid.UUID{},
		Exists: false,
	}, nil
}

func (s *ResoService) GetGangOverlap(gangID1, gangID2 uuid.UUID) (*domain.GangOverlapResult, error) {
	commonMembers, err := s.repo.GetGangOverlap(gangID1, gangID2)
	if err != nil {
		return nil, err
	}

	gangs, err := s.repo.GetAllGangs()
	if err != nil {
		return nil, err
	}

	gangMap := make(map[uuid.UUID]domain.Gang)
	for _, g := range gangs {
		gangMap[g.GangID] = g
	}

	result := &domain.GangOverlapResult{
		GangID1:       gangID1,
		GangID2:       gangID2,
		CommonMembers: commonMembers,
		OverlapCount:  len(commonMembers),
	}

	if g1, ok := gangMap[gangID1]; ok {
		result.Gang1Name = g1.Name
	}
	if g2, ok := gangMap[gangID2]; ok {
		result.Gang2Name = g2.Name
	}

	relations, err := s.repo.GetGangRelations()
	if err == nil {
		for _, rel := range relations {
			if (rel.GangID1 == gangID1 && rel.GangID2 == gangID2) ||
				(rel.GangID1 == gangID2 && rel.GangID2 == gangID1) {
				result.RelationType = &rel.RelationType
				break
			}
		}
	}

	return result, nil
}

func (s *ResoService) DetectCommunities() (*domain.CommunityDetectionResult, error) {
	associations, err := s.repo.GetPersonAssociations()
	if err != nil {
		return nil, err
	}

	adj := make(map[uuid.UUID]map[uuid.UUID]float64)
	for _, a := range associations {
		if adj[a.PersonID1] == nil {
			adj[a.PersonID1] = make(map[uuid.UUID]float64)
		}
		if adj[a.PersonID2] == nil {
			adj[a.PersonID2] = make(map[uuid.UUID]float64)
		}
		adj[a.PersonID1][a.PersonID2] = a.Confidence
		adj[a.PersonID2][a.PersonID1] = a.Confidence
	}

	communities := labelPropagation(adj)
	modularity := computeModularity(adj, communities)

	result := &domain.CommunityDetectionResult{
		Communities:   communities,
		Modularity:    modularity,
		TotalClusters: len(communities),
	}

	return result, nil
}

func labelPropagation(adj map[uuid.UUID]map[uuid.UUID]float64) []domain.Community {
	labels := make(map[uuid.UUID]int)
	i := 0
	for id := range adj {
		labels[id] = i
		i++
	}

	for iter := 0; iter < 50; iter++ {
		changed := false
		for id, neighbors := range adj {
			if len(neighbors) == 0 {
				continue
			}
			labelCount := make(map[int]float64)
			for neighborID, weight := range neighbors {
				if lbl, ok := labels[neighborID]; ok {
					labelCount[lbl] += weight
				}
			}
			maxCount := -1.0
			bestLabel := labels[id]
			for lbl, count := range labelCount {
				if count > maxCount {
					maxCount = count
					bestLabel = lbl
				}
			}
			if bestLabel != labels[id] {
				labels[id] = bestLabel
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	groups := make(map[int][]uuid.UUID)
	for id, lbl := range labels {
		groups[lbl] = append(groups[lbl], id)
	}

	communities := make([]domain.Community, 0, len(groups))
	id := 0
	for _, members := range groups {
		if len(members) < 2 {
			continue
		}
		communities = append(communities, domain.Community{
			ID:      id,
			Members: members,
			Size:    len(members),
		})
		id++
	}

	return communities
}

func computePageRank(adj map[uuid.UUID]map[uuid.UUID]float64) map[uuid.UUID]float64 {
	damping := 0.85
	n := float64(len(adj))
	scores := make(map[uuid.UUID]float64)
	newScores := make(map[uuid.UUID]float64)

	for id := range adj {
		scores[id] = 1.0 / n
	}

	for iter := 0; iter < 50; iter++ {
		for id := range adj {
			newScores[id] = (1 - damping) / n
		}
		for id, neighbors := range adj {
			outDegree := float64(len(neighbors))
			if outDegree == 0 {
				continue
			}
			share := scores[id] / outDegree
			for neighborID := range neighbors {
				newScores[neighborID] += damping * share
			}
		}
		scores, newScores = newScores, scores
	}

	return scores
}

func computeBetweenness(adj map[uuid.UUID]map[uuid.UUID]float64) map[uuid.UUID]float64 {
	betweenness := make(map[uuid.UUID]float64)
	for id := range adj {
		betweenness[id] = 0
	}

	for source := range adj {
		stack := []uuid.UUID{}
		predecessors := make(map[uuid.UUID][]uuid.UUID)
		sigma := make(map[uuid.UUID]float64)
		delta := make(map[uuid.UUID]float64)
		dist := make(map[uuid.UUID]int)

		for id := range adj {
			sigma[id] = 0
			dist[id] = -1
			delta[id] = 0
		}
		sigma[source] = 1
		dist[source] = 0

		queue := []uuid.UUID{source}
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)
			for w := range adj[v] {
				if dist[w] < 0 {
					dist[w] = dist[v] + 1
					queue = append(queue, w)
				}
				if dist[w] == dist[v]+1 {
					sigma[w] += sigma[v]
					predecessors[w] = append(predecessors[w], v)
				}
			}
		}

		for i := len(stack) - 1; i >= 0; i-- {
			w := stack[i]
			for _, v := range predecessors[w] {
				delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
			}
			if w != source {
				betweenness[w] += delta[w]
			}
		}
	}

	maxBC := 0.0
	for _, v := range betweenness {
		if v > maxBC {
			maxBC = v
		}
	}
	if maxBC > 0 {
		for id := range betweenness {
			betweenness[id] /= maxBC
		}
	}

	return betweenness
}

func computeDegreeCentrality(adj map[uuid.UUID]map[uuid.UUID]float64) map[uuid.UUID]float64 {
	scores := make(map[uuid.UUID]float64)
	maxDegree := 0.0

	for id, neighbors := range adj {
		degree := float64(len(neighbors))
		scores[id] = degree
		if degree > maxDegree {
			maxDegree = degree
		}
	}

	if maxDegree > 0 {
		for id := range scores {
			scores[id] /= maxDegree
		}
	}

	return scores
}

func computeModularity(adj map[uuid.UUID]map[uuid.UUID]float64, communities []domain.Community) float64 {
	totalWeight := 0.0
	for _, neighbors := range adj {
		for _, w := range neighbors {
			totalWeight += w
		}
	}
	totalWeight /= 2

	if totalWeight == 0 {
		return 0
	}

	m := 2 * totalWeight
	Q := 0.0

	nodeCommunity := make(map[uuid.UUID]int)
	for _, comm := range communities {
		for _, member := range comm.Members {
			nodeCommunity[member] = comm.ID
		}
	}

	for _, comm := range communities {
		sumIn := 0.0
		sumDegree := 0.0
		for _, i := range comm.Members {
			for j, w := range adj[i] {
				if nodeCommunity[j] == comm.ID {
					sumIn += w
				}
			}
			sumDegree += float64(len(adj[i]))
		}
		Q += sumIn/m - math.Pow(sumDegree/m, 2)
	}

	return Q
}
