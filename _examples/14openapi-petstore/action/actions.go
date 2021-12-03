package action

import (
	"m/14openapi-petstore/design"
	"sync"

	"github.com/morikuni/failure"
)

type PetStore struct {
	Pets   map[int64]*Pet
	NextId int64
	Lock   sync.Mutex
}

// NewPet defines model for NewPet.
type NewPet struct {
	// Name of the pet
	Name string `json:"name"`

	// Type of the pet
	Tag *string `json:"tag,omitempty"`
}

// Pet defines model for Pet.
type Pet struct {
	// Embedded struct due to allOf(#/components/schemas/NewPet)
	NewPet `yaml:",inline"`
	// Embedded fields due to inline allOf schema
	// Unique id of the pet
	Id int64 `json:"id"`
}

func FindPets(
	p *PetStore,
	tags *[]string, limit *int32,
) ([]*Pet, error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var result []*Pet

	for _, pet := range p.Pets {
		if tags != nil {
			// If we have tags,  filter pets by tag
			for _, t := range *tags {
				if pet.Tag != nil && (*pet.Tag == t) {
					result = append(result, pet)
				}
			}
		} else {
			// Add all pets if we're not filtering
			result = append(result, pet)
		}

		if limit != nil {
			l := int(*limit)
			if len(result) >= l {
				// We're at the limit
				break
			}
		}
	}
	return result, nil
}

func AddPet(
	p *PetStore,
	params NewPet,
) (*Pet, error) {
	// We now have a pet, let's add it to our "database".

	// We're always asynchronous, so lock unsafe operations below
	p.Lock.Lock()
	defer p.Lock.Unlock()

	// We handle pets, not NewPets, which have an additional ID field
	pet := new(Pet)
	pet.Name = params.Name
	pet.Tag = params.Tag
	pet.Id = p.NextId
	p.NextId = p.NextId + 1

	// Insert into map
	p.Pets[pet.Id] = pet
	return pet, nil
}

func FindPetByID(
	p *PetStore,
	id int64,
) (*Pet, error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[id]
	if !found {
		return nil, failure.New(design.CodeNotFound, failure.Messagef("Could not find pet with ID %d", id))
	}

	return pet, nil
}

func DeletePet(
	p *PetStore,
	id int64,
) (interface{}, error) { // TODO: 204
	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[id]
	if !found {
		return nil, failure.New(design.CodeNotFound, failure.Messagef("Could not find pet with ID %d", id))
	}
	delete(p.Pets, id)
	return nil, nil
}
