package tests

import (
	"fmt"
	"testing"

	"github.com/badico-cloud-hub/pubsub/dto"
	"github.com/badico-cloud-hub/pubsub/helpers"
	"github.com/joho/godotenv"
)

func TestValidatePropsJwtDto(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestValidatePropsJwtDto: expect(nil) got(%s)\n", err.Error())
	}
	jwtDto := dto.JWTDTO{
		Payload: dto.JWTDTOPayload{
			ClientId:      "000001@orion",
			AssociationId: "pix@orion",
			ApiKeyType:    "pix",
			Provider:      "starkbank",
		},
	}
	if err := helpers.ValidatePayloadJwtDto(jwtDto); err != nil {
		t.Errorf("TestValidatePropsJwtDto: expect(nil) got(%s)\n", err.Error())
	}
}

func TestVerifyBearerToken(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Errorf("TestVerifyBearerToken: expect(nil) got(%s)\n", err.Error())
	}
	bearer := "Bearer eyJhbGciOiJSU0EtT0FFUCIsImVuYyI6IkEyNTZHQ00ifQ.nOQY9gH4Cy0W0AT4bQyVWSCxAIskJYZy_RPpGyxWfyWYUMJzpmQYwcuGw-GLTYcpgu0bTWBlMx4PzsROdqPse4tD20xPtiIKhoJqwzh40WvFOZ6RDUAxAkq1A_PxSA4D_7wln9Uk9o6t-Enx4QUf4RI6q2WtVg9LhRzgbyHn_mak6xoUguYssXHOpy3nA6LqN4wXfZzwUj8QAs0JGuk0wLKtRcTHUg6KIhuU6gHBJtyYZ3h_8MfAa1pnZW95BEbDAuR86PJdv3cSaGOXNkri_kKVf0KDfiRARUwu46zJ2oyOLCKGVzwVSA9qbnDwdZtVsUHyao1nKLrjfSujaIQPdmC8U40584vZU-BUlFBLZJ62iNjRi_aKLJCJiXLETzakqlQcoOdgTp-3SjDejdGgoVCx4NLaVgXw83qsxrIWRIR8jQF4BCLD9eLK4-GLsbkWQknXhd8ph8gYoF3KCIuc7RsOHuI_ZFtnqiHExPsQrPv-DClTAy3g3QUPCsb23lVQ.34bnF1EGDO01Szp-.FDQVQuv1wKrOccC_vYzB5G_scz_OS8X7UwDJywBHceGtTWF96wGkrPJVB8gwjqbZxKyhPntI6_Ck8o4vrvTaciRvH3k1RYmcrv7Qm1rQKXnI8noHVj95t17BSf6AXm59vXvdAXmoItlrYMN2ddn5P2OAC0H-3X4NxfJiqZO20ivk9UqxQsillO-MHTFiBwXJfUWtBQAFuUIvmJbQFNSOpaL-t19iRbtX2rGhug-H9elMclEyvrgpLpSah2IvF0nkAqQQDyXArlOQ2wdBED6wBwa-6aFu87W82nTYtBhlRWSCEb5RlI8NxjcnZ8RoYPM0VhnPs11XK2bGrN_u_tfoTO9z0dQzCIFRLF5-Ojht8gus-O3D9T3VTk8t92Ojg-NPa9Sif9vxhonOwceayA7sIC6nYLp6q6heB7XAUUODa9P3OOxfkgF2BsA4BAGbPPM2y3OkQQ.canyfSDvgGB55X1oJeY4mA"
	jwtDto, err := helpers.VerifyToken(bearer)
	if err != nil {
		t.Errorf("TestVerifyBearerToken: expect(nil) got(%s)\n", err.Error())
	}
	fmt.Printf("jwtDto: %+v\n", jwtDto)
}
